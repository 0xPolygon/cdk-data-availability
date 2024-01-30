package sequencer

import (
	"context"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/polygonvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/common"
)

// Tracker watches the contract for relevant changes to the sequencer
type Tracker struct {
	client    etherman.Etherman
	stop      chan struct{}
	timeout   time.Duration
	retry     time.Duration
	addr      common.Address
	url       string
	lock      sync.Mutex
	startOnce sync.Once
}

// NewTracker creates a new Tracker
func NewTracker(cfg config.L1Config, ethClient etherman.Etherman) (*Tracker, error) {
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
	w := &Tracker{
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
func (st *Tracker) GetAddr() common.Address {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.addr
}

func (st *Tracker) setAddr(addr common.Address) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.addr = addr
}

// GetUrl returns the last known URL of the Sequencer
func (st *Tracker) GetUrl() string {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.url
}

func (st *Tracker) setUrl(url string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.url = url
}

// Start starts the SequencerTracker
func (st *Tracker) Start(ctx context.Context) {
	st.startOnce.Do(func() {
		go st.trackAddrChanges(ctx)
		go st.trackUrlChanges(ctx)
	})
}

func (st *Tracker) trackAddrChanges(ctx context.Context) {
	events := make(chan *polygonvalidium.PolygonvalidiumSetTrustedSequencer)
	defer close(events)

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("context cancelled: %v", ctx.Err())
			}
		default:
			ctx, cancel := context.WithTimeout(ctx, st.timeout)

			sub, err := st.client.WatchSetTrustedSequencer(ctx, events)

			// if no subscription, retry until established
			for err != nil {
				<-time.After(st.retry)

				if sub, err = st.client.WatchSetTrustedSequencer(ctx, events); err != nil {
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
}

func (st *Tracker) trackUrlChanges(ctx context.Context) {
	events := make(chan *polygonvalidium.PolygonvalidiumSetTrustedSequencerURL)
	defer close(events)

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("context cancelled: %v", ctx.Err())
			}
		default:
			ctx, cancel := context.WithTimeout(ctx, st.timeout)

			sub, err := st.client.WatchSetTrustedSequencerURL(ctx, events)

			// if no subscription, retry until established
			for err != nil {
				<-time.After(st.retry)

				if sub, err = st.client.WatchSetTrustedSequencerURL(ctx, events); err != nil {
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
}

// GetSequenceBatch returns sequence batch for given batch number
func (st *Tracker) GetSequenceBatch(batchNum uint64) (*SeqBatch, error) {
	return GetData(st.GetUrl(), batchNum)
}

// Stop stops the SequencerTracker
func (st *Tracker) Stop() {
	close(st.stop)
}
