package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/test/operations"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestSequencerAddrExists(t *testing.T) {

	err := operations.StartComponent("network")
	require.NoError(t, err)
	defer operations.StopComponent("network")
	<-time.After(3 * time.Second) // wait for component to start

	ctx := cli.NewContext(cli.NewApp(), nil, nil)
	cfg, err := config.Load(ctx)
	require.NoError(t, err)
	etherman, err := etherman.New(cfg.L1)
	require.NoError(t, err)

	tracker, err := sequencer.NewSequencerTracker(cfg.L1, etherman)
	require.NoError(t, err)

	go tracker.Start()
	defer tracker.Stop()

	addr := tracker.GetAddr()
	require.Equal(t, common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), addr)

	url := tracker.GetUrl()
	require.Equal(t, "http://cdk-validium-json-rpc:8123", url) // the default

	clientL1, err := ethclient.Dial(operations.DefaultL1NetworkURL)
	require.NoError(t, err)
	validiumContract, err := cdkvalidium.NewCdkvalidium(
		common.HexToAddress(operations.DefaultL1CDKValidiumSmartContract),
		clientL1,
	)
	require.NoError(t, err)

	authL1, err := operations.GetAuth(operations.DefaultSequencerPrivateKey, operations.DefaultL1ChainID)
	require.NoError(t, err)

	newUrl := fmt.Sprintf("http://something-else:%d", rand.Intn(10000))
	_, err = validiumContract.SetTrustedSequencerURL(authL1, newUrl)
	require.NoError(t, err)

	// give the tracker a sec to get the event
	<-time.After(2500 * time.Millisecond)

	require.Equal(t, newUrl, tracker.GetUrl())
}
