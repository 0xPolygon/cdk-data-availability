package synchronizer

import (
	"context"
	"time"

	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/0xPolygon/supernets2-node/etherman"
	"github.com/0xPolygon/supernets2-node/etherman/smartcontracts/supernets2"
	"github.com/0xPolygon/supernets2-node/etherman/smartcontracts/supernets2datacommittee"
	"github.com/0xPolygon/supernets2-node/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// watcher is the base struct for components that watch events on chain. These watchers must only use free RPC calls.
type watcher struct {
	client  *etherman.Client
	stop    chan struct{}
	timeout time.Duration
	retry   time.Duration
}

func newWatcher(config config.L1Config) (*watcher, error) {
	client, err := newEtherman(config)
	if err != nil {
		return nil, err
	}
	return &watcher{
		client:  client,
		stop:    make(chan struct{}),
		timeout: config.Timeout.Duration,
		retry:   config.RetryPeriod.Duration,
	}, nil
}

// newEtherman constructs an etherman client that only needs the free contract calls
func newEtherman(cfg config.L1Config) (*etherman.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.Duration)
	defer cancel()
	ethClient, err := ethclient.DialContext(ctx, cfg.WsURL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.WsURL, err)
		return nil, err
	}
	supernets2, err := supernets2.NewSupernets2(common.HexToAddress(cfg.Supernets2Address), ethClient)
	if err != nil {
		return nil, err
	}
	dataCommittee, err :=
		supernets2datacommittee.NewSupernets2datacommittee(common.HexToAddress(cfg.DataCommitteeAddress), ethClient)
	if err != nil {
		return nil, err
	}
	return &etherman.Client{
		EthClient:     ethClient,
		Supernets2:    supernets2,
		DataCommittee: dataCommittee,
	}, nil
}

func handleSubscriptionContextDone(ctx context.Context) {
	// Deadline exceeded is expected since we use finite context timeout
	if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
		log.Warnf("re-establishing subscription: %v", ctx.Err())
	}
}
