package synchronizer

import (
	"context"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-validium-node/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-validium-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

// SequencerTracker watches the contract for relevant changes to the sequencer
type SequencerTracker struct {
	watcher
	addr common.Address
	lock sync.Mutex
}

// NewSequencerTracker creates a new SequencerTracker
func NewSequencerTracker(cfg config.L1Config) (*SequencerTracker, error) {
	log.Info("starting sequencer tracker")
	watcher, err := newWatcher(cfg)
	if err != nil {
		return nil, err
	}
	// current address of the sequencer
	addr, err := watcher.client.TrustedSequencer()
	if err != nil {
		return nil, err
	}
	w := &SequencerTracker{
		watcher: *watcher,
		addr:    addr,
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

// Start starts the SequencerTracker
func (st *SequencerTracker) Start() {
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
			handleSubscriptionContextDone(ctx)
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
