package e2e

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/etherman"
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

	clientL1, err := ethclient.Dial(operations.DefaultL1NetworkURL)
	require.NoError(t, err)

	validiumContract, err := polygonvalidiumetrog.NewPolygonvalidiumetrog(
		common.HexToAddress(operations.DefaultL1CDKValidiumSmartContract),
		clientL1,
	)
	require.NoError(t, err)

	authL1, err := operations.GetAuth(operations.DefaultSequencerPrivateKey, operations.DefaultL1ChainID)
	require.NoError(t, err)

	newUrl := fmt.Sprintf("http://something-else:%d", rand.Intn(10000))

	initTracker := func(rpcUrl string) *sequencer.Tracker {
		cfg, err := config.Load(ctx)
		require.NoError(t, err)

		// Make sure ws is used
		cfg.L1.RpcURL = rpcUrl
		cfg.L1.TrackSequencerPollInterval = types.NewDuration(100 * time.Millisecond)

		etm, err := etherman.New(ctx.Context, cfg.L1)
		require.NoError(t, err)

		tracker := sequencer.NewTracker(cfg.L1, etm)

		tracker.Start(ctx.Context)

		addr := tracker.GetAddr()
		require.Equal(t, common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), addr)

		url := tracker.GetUrl()
		require.Equal(t, "http://zkevm-json-rpc:8123", url) // the default

		return tracker
	}

	wsTracker := initTracker("ws://127.0.0.1:8546")
	defer wsTracker.Stop()

	httpTracker := initTracker("http://127.0.0.1:8545")
	defer httpTracker.Stop()

	// Update URL on L1 contract
	_, err = validiumContract.SetTrustedSequencerURL(authL1, newUrl)
	require.NoError(t, err)

	// Give the tracker a sec to get the event
	<-time.After(2500 * time.Millisecond)

	require.Equal(t, newUrl, wsTracker.GetUrl())
	require.Equal(t, newUrl, httpTracker.GetUrl())
}
