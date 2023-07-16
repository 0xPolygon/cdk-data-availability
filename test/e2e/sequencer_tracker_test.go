package e2e

import (
	"testing"
	"time"

	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestTracker(t *testing.T) {
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
