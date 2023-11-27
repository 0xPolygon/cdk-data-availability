package sequencer

import (
	"context"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// SequencerTracker watches the contract for relevant changes to the sequencer
type SequencerTracker struct {
	client  *etherman.Etherman
	stop    chan struct{}
	timeout time.Duration
	retry   time.Duration
	addr    common.Address
	url     string
	lock    sync.Mutex
}

// NewSequencerTracker creates a new SequencerTracker
func NewSequencerTracker(cfg config.L1Config, ethClient *etherman.Etherman) (*SequencerTracker, error) {
	log.Info("starting sequencer address tracker")
	addr, err := ethClient.TrustedSequencer()
	if err != nil {
		return nil, err
	}
	log.Infof("current sequencer addr: %s", addr.Hex())
	url, err := ethClient.TrustedSequencerURL()
	if err != nil {
		return nil, err
	}
	log.Infof("current sequencer url: %s", url)
	w := &SequencerTracker{
		client:  ethClient,
		stop:    make(chan struct{}),
		timeout: cfg.Timeout.Duration,
		retry:   cfg.RetryPeriod.Duration,
		addr:    addr,
		url:     url,
	}
	return w, nil
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

// GetUrl returns the last known URL of the Sequencer
func (st *SequencerTracker) GetUrl() string {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.url
}

func (st *SequencerTracker) setUrl(url string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.url = url
}

// Start starts the SequencerTracker
func (st *SequencerTracker) Start() {
	go st.trackAddrChanges()
	go st.trackUrlChanges()
}

func (st *SequencerTracker) trackAddrChanges() {
	events := make(chan *cdkvalidium.CdkvalidiumSetTrustedSequencer)
	defer close(events)
	for {
		var (
			sub event.Subscription
			err error
		)

		ctx, cancel := context.WithTimeout(context.Background(), st.timeout)
		opts := &bind.WatchOpts{Context: ctx}
		sub, err = st.client.CDKValidium.WatchSetTrustedSequencer(opts, events)

		// if no subscription, retry until established
		for err != nil {
			<-time.After(st.retry)
			sub, err = st.client.CDKValidium.WatchSetTrustedSequencer(opts, events)
			if err != nil {
				log.Errorf("error subscribing to trusted sequencer event, retrying: %v", err)
			}
		}

		// wait on events, timeouts, and signals to stop
		select {
		case e := <-events:
			log.Infof("new trusted sequencer address: %v", e.NewTrustedSequencer)
			st.setAddr(e.NewTrustedSequencer)
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing: %v", err)
		case <-ctx.Done():
			// Deadline exceeded is expected since we use finite context timeout
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("re-establishing subscription: %v", ctx.Err())
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

func (st *SequencerTracker) trackUrlChanges() {
	events := make(chan *cdkvalidium.CdkvalidiumSetTrustedSequencerURL)
	defer close(events)
	for {
		var (
			sub event.Subscription
			err error
		)

		ctx, cancel := context.WithTimeout(context.Background(), st.timeout)
		opts := &bind.WatchOpts{Context: ctx}
		sub, err = st.client.CDKValidium.WatchSetTrustedSequencerURL(opts, events)

		// if no subscription, retry until established
		for err != nil {
			<-time.After(st.retry)
			sub, err = st.client.CDKValidium.WatchSetTrustedSequencerURL(opts, events)
			if err != nil {
				log.Errorf("error subscribing to trusted sequencer event, retrying: %v", err)
			}
		}

		// wait on events, timeouts, and signals to stop
		select {
		case e := <-events:
			log.Infof("new trusted sequencer url: %v", e.NewTrustedSequencerURL)
			st.setUrl(e.NewTrustedSequencerURL)
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing: %v", err)
		case <-ctx.Done():
			// Deadline exceeded is expected since we use finite context timeout
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("re-establishing subscription: %v", ctx.Err())
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
