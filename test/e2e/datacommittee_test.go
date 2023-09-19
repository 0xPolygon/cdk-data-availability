package e2e

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	cTypes "github.com/0xPolygon/cdk-validium-node/config/types"
	"github.com/0xPolygon/cdk-validium-node/db"
	"github.com/0xPolygon/cdk-validium-node/etherman/smartcontracts/cdkdatacommittee"
	"github.com/0xPolygon/cdk-validium-node/jsonrpc"
	"github.com/0xPolygon/cdk-validium-node/log"
	"github.com/0xPolygon/cdk-validium-node/test/operations"
	"github.com/ethereum/go-ethereum"
	eTypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nSignatures      = 4
	mMembers         = 5
	ksFile           = "/tmp/pkey"
	cfgFile          = "/tmp/dacnodeconfigfile.json"
	ksPass           = "pass"
	dacNodeContainer = "hermeznetwork/cdk-data-availability:v0.0.1"
)

func TestDataCommittee(t *testing.T) {
	// Setup
	var err error
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	defer func() {
		require.NoError(t, operations.Teardown())
	}()
	err = operations.Teardown()
	require.NoError(t, err)
	opsCfg := operations.GetDefaultOperationsConfig()
	opsCfg.State.MaxCumulativeGasUsed = 80000000000
	opsman, err := operations.NewManager(ctx, opsCfg)
	require.NoError(t, err)
	err = opsman.Setup()
	require.NoError(t, err)
	time.Sleep(5 * time.Second)
	authL2, err := operations.GetAuth(operations.DefaultSequencerPrivateKey, operations.DefaultL2ChainID)
	require.NoError(t, err)
	authL1, err := operations.GetAuth(operations.DefaultSequencerPrivateKey, operations.DefaultL1ChainID)
	require.NoError(t, err)
	clientL2, err := ethclient.Dial(operations.DefaultL2NetworkURL)
	require.NoError(t, err)
	clientL1, err := ethclient.Dial(operations.DefaultL1NetworkURL)
	require.NoError(t, err)
	dacSC, err := cdkdatacommittee.NewCdkdatacommittee(
		common.HexToAddress(operations.DefaultL1DataCommitteeContract),
		clientL1,
	)
	require.NoError(t, err)

	// Register committe with N / M signatures
	membs := members{}
	addrsBytes := []byte{}
	urls := []string{}
	for i := 0; i < mMembers; i++ {
		pk, err := crypto.GenerateKey()
		require.NoError(t, err)
		membs = append(membs, member{
			addr: crypto.PubkeyToAddress(pk.PublicKey),
			pk:   pk,
			url:  fmt.Sprintf("http://cdk-data-availability-%d:420%d", i, i),
			i:    i,
		})
	}
	sort.Sort(membs)
	for _, m := range membs {
		addrsBytes = append(addrsBytes, m.addr.Bytes()...)
		urls = append(urls, m.url)
	}
	tx, err := dacSC.SetupCommittee(authL1, big.NewInt(nSignatures), urls, addrsBytes)
	if err != nil {
		for _, m := range membs {
			fmt.Println(m.addr)
		}
	}
	require.NoError(t, err)
	err = operations.WaitTxToBeMined(ctx, clientL1, tx, operations.DefaultTimeoutTxToBeMined)
	require.NoError(t, err)

	defer func() {
		// Remove tmp files
		assert.NoError(t,
			exec.Command("rm", cfgFile).Run(),
		)
		assert.NoError(t,
			exec.Command("rmdir", ksFile+"_").Run(),
		)
		assert.NoError(t,
			exec.Command("rm", ksFile).Run(),
		)
		// Stop the members
		for _, m := range membs {
			stopDACMember(t, m)
		}

		// Stop permissionless node
		require.NoError(t, opsman.StopPermissionlessNodeForcedToSYncThroughDAC())
	}()

	// Start permissionless node
	require.NoError(t, opsman.StartPermissionlessNodeForcedToSYncThroughDAC())

	// Start DAC nodes & DBs
	for _, m := range membs {
		startDACMember(t, m)
	}

	// Send txs
	nTxs := 10
	amount := big.NewInt(10000)
	toAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	_, err = clientL2.BalanceAt(ctx, authL2.From, nil)
	require.NoError(t, err)
	_, err = clientL2.PendingNonceAt(ctx, authL2.From)
	require.NoError(t, err)

	gasLimit, err := clientL2.EstimateGas(ctx, ethereum.CallMsg{From: authL2.From, To: &toAddress, Value: amount})
	require.NoError(t, err)

	gasPrice, err := clientL2.SuggestGasPrice(ctx)
	require.NoError(t, err)

	nonce, err := clientL2.PendingNonceAt(ctx, authL2.From)
	require.NoError(t, err)

	txs := make([]*eTypes.Transaction, 0, nTxs)
	for i := 0; i < nTxs; i++ {
		tx := eTypes.NewTransaction(nonce+uint64(i), toAddress, amount, gasLimit, gasPrice, nil)
		log.Infof("generating tx %d / %d: %s", i+1, nTxs, tx.Hash().Hex())
		txs = append(txs, tx)
	}

	// Wait for verification
	_, err = operations.ApplyL2Txs(ctx, txs, authL2, clientL2, operations.VerifiedConfirmationLevel)
	require.NoError(t, err)

	// Assert that the permissionless node is fully synced (through the DAC)
	time.Sleep(30 * time.Second) // Give some time for the permissionless node to get synced
	clientL2Permissionless, err := ethclient.Dial(operations.PermissionlessL2NetworkURL)
	require.NoError(t, err)
	expectedBlock, err := clientL2.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	actualBlock, err := clientL2Permissionless.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	// je, err := expectedBlock.Header().MarshalJSON()
	// require.NoError(t, err)
	// log.Info(string(je))
	// ja, err := actualBlock.Header().MarshalJSON()
	// require.NoError(t, err)
	// log.Info(string(ja))
	// require.Equal(t, string(je), string(ja))
	require.Equal(t, expectedBlock.Root().Hex(), actualBlock.Root().Hex())
}

type member struct {
	addr common.Address
	pk   *ecdsa.PrivateKey
	url  string
	i    int
}
type members []member

func (s members) Len() int { return len(s) }
func (s members) Less(i, j int) bool {
	return strings.ToUpper(s[i].addr.Hex()) < strings.ToUpper(s[j].addr.Hex())
}
func (s members) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func createKeyStore(pk *ecdsa.PrivateKey, outputDir, password string) error {
	ks := keystore.NewKeyStore(outputDir+"_", keystore.StandardScryptN, keystore.StandardScryptP)
	_, err := ks.ImportECDSA(pk, password)
	if err != nil {
		return err
	}
	fileNameB, err := exec.Command("ls", outputDir+"_/").CombinedOutput()
	fileName := strings.TrimSuffix(string(fileNameB), "\n")
	if err != nil {
		fmt.Println(fileName)
		return err
	}
	out, err := exec.Command("mv", outputDir+"_/"+fileName, outputDir).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return err
	}
	return nil
}

func startDACMember(t *testing.T, m member) {
	dacNodeConfig := config.Config{
		L1: config.L1Config{
			RpcURL:               "http://cdk-validium-mock-l1-network:8545",
			WsURL:                "ws://cdk-validium-mock-l1-network:8546",
			CDKValidiumAddress:   operations.DefaultL1CDKValidiumSmartContract,
			DataCommitteeAddress: operations.DefaultL1DataCommitteeContract,
			Timeout:              cTypes.Duration{Duration: time.Second},
			RetryPeriod:          cTypes.Duration{Duration: time.Second},
		},
		PrivateKey: cTypes.KeystoreFileConfig{
			Path:     ksFile,
			Password: ksPass,
		},
		DB: db.Config{
			Name:      "committee_db",
			User:      "committee_user",
			Password:  "committee_password",
			Host:      "cdk-validium-data-node-db-" + strconv.Itoa(m.i),
			Port:      "5432",
			EnableLog: false,
			MaxConns:  10,
		},
		RPC: jsonrpc.Config{
			Host:                             "0.0.0.0",
			EnableL2SuggestedGasPricePolling: false,
			MaxRequestsPerIPAndSecond:        100,
		},
	}

	// Run the DB
	dbCmd := exec.Command(
		"docker", "run", "-d",
		"--name", dacNodeConfig.DB.Host,
		"-e", "POSTGRES_DB=committee_db",
		"-e", "POSTGRES_PASSWORD=committee_password",
		"-e", "POSTGRES_USER=committee_user",
		"-p", fmt.Sprintf("553%d:5432", m.i),
		"--network", "cdk-validium",
		"postgres", "-N", "500",
	)
	out, err := dbCmd.CombinedOutput()
	require.NoError(t, err, string(out))
	log.Infof("DAC DB %d started", m.i)
	time.Sleep(time.Second * 2)

	// Set correct port
	port := 4200 + m.i
	dacNodeConfig.RPC.Port = port

	// Write config file
	file, err := json.MarshalIndent(dacNodeConfig, "", " ")
	require.NoError(t, err)
	err = os.WriteFile(cfgFile, file, 0644)
	require.NoError(t, err)
	// Write private key keystore file
	err = createKeyStore(m.pk, ksFile, ksPass)
	require.NoError(t, err)
	// Run DAC node
	cmd := exec.Command(
		"docker", "run", "-d",
		"--name", "cdk-data-availability-"+strconv.Itoa(m.i),
		"-v", cfgFile+":/app/config.json",
		"-v", ksFile+":"+ksFile,
		"--network", "cdk-validium",
		dacNodeContainer,
		"/bin/sh", "-c",
		"/app/cdk-data-availability run --cfg /app/config.json",
	)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	log.Infof("DAC node %d started", m.i)
	time.Sleep(time.Second * 5)
}

func stopDACMember(t *testing.T, m member) {
	assert.NoError(t, exec.Command(
		"docker", "kill", "cdk-data-availability-"+strconv.Itoa(m.i),
	).Run())
	assert.NoError(t, exec.Command(
		"docker", "rm", "cdk-data-availability-"+strconv.Itoa(m.i),
	).Run())
	assert.NoError(t, exec.Command(
		"docker", "kill", "cdk-validium-data-node-db-"+strconv.Itoa(m.i),
	).Run())
	assert.NoError(t, exec.Command(
		"docker", "rm", "cdk-validium-data-node-db-"+strconv.Itoa(m.i),
	).Run())
}
