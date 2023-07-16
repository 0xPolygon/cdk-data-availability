package synchronizer

import (
	"context"
	"sync"
	"time"

	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/0xPolygon/supernets2-node/etherman"
	"github.com/0xPolygon/supernets2-node/etherman/smartcontracts/supernets2datacommittee"
	"github.com/0xPolygon/supernets2-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/event"
)

// DataCommitteeTracker tracks changes to the data committee membership
type DataCommitteeTracker struct {
	watcher
	committee *etherman.DataCommittee
	lock      sync.Mutex
}

// NewDataCommitteeTracker creates the DataCommitteeTracker
func NewDataCommitteeTracker(cfg config.L1Config) (*DataCommitteeTracker, error) {
	watcher, err := newWatcher(cfg)
	if err != nil {
		return nil, err
	}
	// current data committee
	committee, err := watcher.client.GetCurrentDataCommittee()
	if err != nil {
		return nil, err
	}
	w := &DataCommitteeTracker{
		watcher:   *watcher,
		committee: committee,
	}
	return w, nil
}

// GetMembers returns the current data committee memebers
func (dct *DataCommitteeTracker) GetMembers() *etherman.DataCommittee {
	dct.lock.Lock()
	defer dct.lock.Unlock()
	return dct.committee
}

func (dct *DataCommitteeTracker) setDataCommittee(committee *etherman.DataCommittee) {
	dct.lock.Lock()
	defer dct.lock.Unlock()
	dct.committee = committee
}

// Start starts the DataCommitteeTracker
func (dct *DataCommitteeTracker) Start() {
	// watch for changes in the data committee
	events := make(chan *supernets2datacommittee.Supernets2datacommitteeCommitteeUpdated)
	defer close(events)
	for {
		var (
			sub event.Subscription
			err error
		)

		ctx, cancel := context.WithTimeout(context.Background(), dct.timeout)
		opts := &bind.WatchOpts{Context: ctx}
		sub, err = dct.client.DataCommittee.WatchCommitteeUpdated(opts, events)

		// if no subscription, retry until established
		for err != nil {
			<-time.After(dct.retry)
			sub, err = dct.client.DataCommittee.WatchCommitteeUpdated(opts, events)
			if err != nil {
				log.Errorf("error subscribing to data committee update events, retrying. %v", err)
			}
		}

		// wait on events, timeouts, and signals to stop
		select {
		case e := <-events:
			log.Infof("new data committee hash: %v", e.CommitteeHash)
			newCommittee, err := dct.client.GetCurrentDataCommittee()
			if err != nil {
				log.Errorf("error retrieving new data committee: %v", err) // TODO: retry?
			}
			if newCommittee != nil {
				dct.setDataCommittee(newCommittee)
			}
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing: %v", err)
		case <-ctx.Done():
			if ctx.Err() != nil {
				log.Warnf("re-establishing subscription: %v", ctx.Err())
			}
		case <-dct.stop:
			if sub != nil {
				sub.Unsubscribe()
			}
			cancel()
			return
		}
	}
}

// Stop stops the DataCommitteeTracker
func (dct *DataCommitteeTracker) Stop() {
	close(dct.stop)
}
