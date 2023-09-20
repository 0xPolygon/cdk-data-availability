package e2e

import (
	"testing"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/synchronizer"
	"github.com/0xPolygon/cdk-validium-node/test/operations"
	"github.com/ethereum/go-ethereum/common"
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

	tracker, err := synchronizer.NewSequencerTracker(cfg.L1)
	require.NoError(t, err)

	addr := tracker.GetAddr()
	require.Equal(t, common.HexToAddress("0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"), addr)

	go tracker.Start()
	select {
	case <-time.After(1 * time.Second):
		tracker.Stop()
	}
}
