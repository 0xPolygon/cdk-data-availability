package sequencer

import (
	"context"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/pkg/backoff"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/etrog/polygonvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// Tracker watches the contract for relevant changes to the sequencer
type Tracker struct {
	client       etherman.Etherman
	stop         chan struct{}
	timeout      time.Duration
	retry        time.Duration
	addr         common.Address
	url          string
	trackChanges bool
	lock         sync.Mutex
	startOnce    sync.Once
}

// NewTracker creates a new Tracker
func NewTracker(cfg config.L1Config, ethClient etherman.Etherman) *Tracker {
	return &Tracker{
		client:       ethClient,
		stop:         make(chan struct{}),
		timeout:      cfg.Timeout.Duration,
		retry:        cfg.RetryPeriod.Duration,
		trackChanges: cfg.TrackSequencer,
	}
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
func (st *Tracker) Start(parentCtx context.Context) {
	st.startOnce.Do(func() {
		ctx, cancel := context.WithTimeout(parentCtx, st.timeout)
		defer cancel()

		addr, err := st.client.TrustedSequencer(ctx)
		if err != nil {
			log.Fatalf("failed to get sequencer addr: %v", err)
			return
		}

		log.Infof("current sequencer addr: %s", addr.Hex())
		st.setAddr(addr)

		url, err := st.client.TrustedSequencerURL(ctx)
		if err != nil {
			log.Fatalf("failed to get sequencer addr: %v", err)
			return
		}

		log.Infof("current sequencer url: %s", url)
		st.setUrl(url)

		if st.trackChanges {
			log.Info("sequencer tracking enabled")

			go st.trackAddrChanges(parentCtx)
			go st.trackUrlChanges(parentCtx)
		}
	})
}

func (st *Tracker) trackAddrChanges(ctx context.Context) {
	events := make(chan *polygonvalidium.PolygonvalidiumSetTrustedSequencer)
	defer close(events)

	var sub event.Subscription

	initSubscription := func() {
		if err := backoff.Exponential(func() (err error) {
			if sub, err = st.client.WatchSetTrustedSequencer(ctx, events); err != nil {
				log.Errorf("error subscribing to trusted sequencer event, retrying: %v", err)
			}

			return err
		}, 5, st.retry); err != nil {
			log.Fatalf("failed subscribing to trusted sequencer event: %v. Check ws(s) availability.", err)
		}
	}

	initSubscription()

	for {
		select {
		case e := <-events:
			log.Infof("new trusted sequencer address: %v", e.NewTrustedSequencer)
			st.setAddr(e.NewTrustedSequencer)
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("context cancelled: %v", ctx.Err())
			}
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing: %v", err)
			initSubscription()
		case <-st.stop:
			if sub != nil {
				sub.Unsubscribe()
			}
			return
		}
	}
}

func (st *Tracker) trackUrlChanges(ctx context.Context) {
	events := make(chan *polygonvalidium.PolygonvalidiumSetTrustedSequencerURL)
	defer close(events)

	var sub event.Subscription

	initSubscription := func() {
		if err := backoff.Exponential(func() (err error) {
			if sub, err = st.client.WatchSetTrustedSequencerURL(ctx, events); err != nil {
				log.Errorf("error subscribing to trusted sequencer URL event, retrying: %v", err)
			}

			return err
		}, 5, st.retry); err != nil {
			log.Fatalf("failed subscribing to trusted sequencer URL event: %v. Check ws(s) availability.", err)
		}
	}

	initSubscription()

	for {
		select {
		case e := <-events:
			log.Infof("new trusted sequencer url: %v", e.NewTrustedSequencerURL)
			st.setUrl(e.NewTrustedSequencerURL)
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("context cancelled: %v", ctx.Err())
			}
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing: %v", err)
			initSubscription()
		case <-st.stop:
			if sub != nil {
				sub.Unsubscribe()
			}
			return
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
