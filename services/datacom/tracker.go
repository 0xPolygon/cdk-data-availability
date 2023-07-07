package datacom

import (
	"context"
	"sync"
	"time"

	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/0xPolygon/supernets2-node/etherman"
	"github.com/0xPolygon/supernets2-node/etherman/smartcontracts/supernets2"
	"github.com/0xPolygon/supernets2-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
)

// SequencerTracker watches the contract for relevant changes to the sequencer
type SequencerTracker struct {
	client  *etherman.Client
	addr    common.Address
	stop    chan struct{}
	lock    sync.Mutex
	timeout time.Duration
	retry   time.Duration
}

// NewSequencerTracker creates a new SequencerTracker
func NewSequencerTracker(cfg config.L1Config) (*SequencerTracker, error) {
	client, err := newEtherman(cfg)
	if err != nil {
		return nil, err
	}
	// current address of the sequencer
	addr, err := client.TrustedSequencer()
	if err != nil {
		return nil, err
	}
	w := &SequencerTracker{
		client:  client,
		addr:    addr,
		stop:    make(chan struct{}),
		timeout: cfg.Timeout.Duration,
		retry:   cfg.RetryPeriod.Duration,
	}
	return w, nil
}

// newEtherman constructs an etherman client that only needs the free API calls to ZkEVMAddr contract
func newEtherman(cfg config.L1Config) (*etherman.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout.Duration)
	defer cancel()
	ethClient, err := ethclient.DialContext(ctx, cfg.WsURL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.WsURL, err)
		return nil, err
	}
	supernets2, err := supernets2.NewSupernets2(common.HexToAddress(cfg.Contract), ethClient)
	if err != nil {
		return nil, err
	}
	return &etherman.Client{
		EthClient:  ethClient,
		Supernets2: supernets2,
	}, nil
}

// GetAddr returns the last known address of the Sequencer
func (st *SequencerTracker) GetAddr() common.Address {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.addr
}

func (st *SequencerTracker) setAddr(addr common.Address) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.addr = addr
}

// Start starts the SequencerTracker
func (st *SequencerTracker) Start() {
	events := make(chan *supernets2.Supernets2SetTrustedSequencer)
	defer close(events)
	for {
		var (
			sub event.Subscription
			err error
		)

		ctx, cancel := context.WithTimeout(context.Background(), st.timeout)
		opts := &bind.WatchOpts{Context: ctx}
		sub, err = st.client.ZkEVM.WatchSetTrustedSequencer(opts, events)

		// if no subscription, retry until established
		for err != nil {
			<-time.After(st.retry)
			sub, err = st.client.ZkEVM.WatchSetTrustedSequencer(opts, events)
			if err != nil {
				log.Errorf("error subscribing to trusted sequencer event, retrying", err)
			}
		}

		// wait on events, timeouts, and signals to stop
		select {
		case e := <-events:
			log.Infof("new trusted sequencer address: {}", e.NewTrustedSequencer)
			st.setAddr(e.NewTrustedSequencer)
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing", err)
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				log.Debug("re-establishing subscription after timeout")
			}
		case <-st.stop:
			if sub != nil {
				sub.Unsubscribe()
			}
			cancel()
			return
		}
	}
}

// Stop stops the SequencerTracker
func (st *SequencerTracker) Stop() {
	close(st.stop)
}
