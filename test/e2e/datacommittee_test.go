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
	cTypes "github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkdatacommittee"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/synchronizer"
	"github.com/0xPolygon/cdk-data-availability/test/operations"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	dacNodeContainer = "cdk-data-availability"
	stopDacs         = true
)

func TestDataCommittee(t *testing.T) {
	// Setup
	var err error
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()
	defer func() {
		if stopDacs {
			require.NoError(t, operations.Teardown())
		}
	}()
	err = operations.Teardown()
	require.NoError(t, err)
	require.NoError(t, err)
	err = operations.Setup()
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

	// The default sequencer URL is incorrect, set it to match the docker container
	validiumContract, err := cdkvalidium.NewCdkvalidium(
		common.HexToAddress(operations.DefaultL1CDKValidiumSmartContract),
		clientL1,
	)
	require.NoError(t, err)
	_, err = validiumContract.SetTrustedSequencerURL(authL1, "http://zkevm-node:8123")
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
		if !stopDacs {
			return
		}
		for _, m := range membs {
			stopDACMember(t, m)
		}
		// Remove tmp files
		assert.NoError(t,
			exec.Command("rm", cfgFile).Run(),
		)
		assert.NoError(t,
			exec.Command("rm", ksFile).Run(),
		)
		// FIXME: for some reason rmdir is failing
		_ = exec.Command("rmdir", "-rf", ksFile+"_").Run()
	}()

	// pick one to start later
	m0 := membs[0]

	// Start DAC nodes & DBs
	for _, m := range membs[1:] { // note starting all but first
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

	startDACMember(t, m0) // start the skipped one, it should catch up through synchronization

	// allow the member to startup and synchronize
	<-time.After(20 * time.Second)

	iter, err := getSequenceBatchesEventIterator(clientL1)
	require.NoError(t, err)
	defer func() { _ = iter.Close() }()

	// All the events should be present in DACs
	for iter.Next() {
		expectedKeys, err := getSequenceBatchesKeys(clientL1, iter.Event)
		require.NoError(t, err)
		for _, m := range membs {
			// Each member (including m0) should have all the keys
			for _, expected := range expectedKeys {
				actual, err := getOffchainDataKeys(m, expected)
				require.NoError(t, err)
				require.Equal(t, expected, actual)
			}
		}
	}
}

func getSequenceBatchesEventIterator(clientL1 *ethclient.Client) (*cdkvalidium.CdkvalidiumSequenceBatchesIterator, error) {
	// Get the expected data keys of the batches from what was submitted to L1
	cdkValidium, err := cdkvalidium.NewCdkvalidium(common.HexToAddress(operations.DefaultL1CDKValidiumSmartContract), clientL1)
	if err != nil {
		return nil, err
	}
	// iterate over all events that were generated
	iter, err := cdkValidium.FilterSequenceBatches(&bind.FilterOpts{Start: 0, Context: context.Background()}, nil)
	if err != nil {
		return nil, err
	}
	return iter, nil
}

func getSequenceBatchesKeys(clientL1 *ethclient.Client, event *cdkvalidium.CdkvalidiumSequenceBatches) ([]common.Hash, error) {
	ctx := context.Background()
	tx, _, err := clientL1.TransactionByHash(ctx, event.Raw.TxHash)
	if err != nil {
		return nil, err
	}
	txData := tx.Data()
	keys, err := synchronizer.UnpackTxData(txData)
	return keys, err
}

func getOffchainDataKeys(m member, tx common.Hash) (common.Hash, error) {
	testUrl := fmt.Sprintf("http://127.0.0.1:420%d", m.i)
	mc := newTestClient(testUrl, m.addr)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	data, err := mc.client.GetOffChainData(ctx, tx)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(data), nil
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
			WsURL:                "ws://l1:8546",
			RpcURL:               "http://l1:8545",
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
		RPC: rpc.Config{
			Host:                      "0.0.0.0",
			MaxRequestsPerIPAndSecond: 100,
		},
		Log: log.Config{
			Level: "debug",
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
		"--network", "cdk-data-availability",
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
		"-p", fmt.Sprintf("%d:%d", port, port),
		"--name", "cdk-data-availability-"+strconv.Itoa(m.i),
		"-v", cfgFile+":/app/config.json",
		"-v", ksFile+":"+ksFile,
		"--network", "cdk-data-availability",
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
	out, err := exec.Command(
		"docker", "kill", "cdk-data-availability-"+strconv.Itoa(m.i),
	).CombinedOutput()
	assert.NoError(t, err, string(out))
	out, err = exec.Command(
		"docker", "rm", "cdk-data-availability-"+strconv.Itoa(m.i),
	).CombinedOutput()
	assert.NoError(t, err, string(out))
	out, err = exec.Command(
		"docker", "kill", "cdk-validium-data-node-db-"+strconv.Itoa(m.i),
	).CombinedOutput()
	assert.NoError(t, err, string(out))
	out, err = exec.Command(
		"docker", "rm", "cdk-validium-data-node-db-"+strconv.Itoa(m.i),
	).CombinedOutput()
	assert.NoError(t, err, string(out))
}
