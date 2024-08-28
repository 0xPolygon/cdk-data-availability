package sequencer

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/pkg/backoff"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
)

const (
	// maxConnectionRetries is the maximum number of retries to connect to the RPC node before failing.
	maxConnectionRetries = 5
)

// Tracker watches the contract for relevant changes to the sequencer
type Tracker struct {
	em           etherman.Etherman
	stop         chan struct{}
	timeout      time.Duration
	retry        time.Duration
	addr         common.Address
	url          string
	trackChanges bool
	usePolling   bool
	pollInterval time.Duration
	wg           sync.WaitGroup
	lock         sync.Mutex
	startOnce    sync.Once
}

// NewTracker creates a new Tracker
func NewTracker(cfg config.L1Config, em etherman.Etherman) *Tracker {
	pollInterval := time.Minute
	if cfg.TrackSequencerPollInterval.Seconds() > 0 {
		pollInterval = cfg.TrackSequencerPollInterval.Duration
	}

	return &Tracker{
		em:           em,
		stop:         make(chan struct{}),
		timeout:      cfg.Timeout.Duration,
		retry:        cfg.RetryPeriod.Duration,
		trackChanges: cfg.TrackSequencer,
		usePolling:   strings.HasPrefix(cfg.RpcURL, "http"), // If http(s), use polling instead of sockets
		pollInterval: pollInterval,
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
	st.addr = addr
	st.lock.Unlock()
}

// GetUrl returns the last known URL of the Sequencer
func (st *Tracker) GetUrl() string {
	st.lock.Lock()
	defer st.lock.Unlock()
	return st.url
}

func (st *Tracker) setUrl(url string) {
	st.lock.Lock()
	st.url = url
	st.lock.Unlock()
}

// Start starts the SequencerTracker
func (st *Tracker) Start(parentCtx context.Context) {
	st.startOnce.Do(func() {
		ctx, cancel := context.WithTimeout(parentCtx, st.timeout)
		defer cancel()

		addr, err := st.em.TrustedSequencer(ctx)
		if err != nil {
			log.Fatalf("failed to get sequencer addr: %v", err)
			return
		}

		log.Infof("current sequencer addr: %s", addr.Hex())
		st.setAddr(addr)

		url, err := st.em.TrustedSequencerURL(ctx)
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
	addrChan := make(chan common.Address, 1)

	if st.usePolling {
		go st.pollAddrChanges(ctx, addrChan)
	} else {
		go st.subscribeOnAddrChanges(ctx, addrChan)
	}

	for {
		select {
		case addr := <-addrChan:
			if st.GetAddr().Cmp(addr) != 0 {
				log.Infof("new trusted sequencer address: %v", addr)
				st.setAddr(addr)
			}
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("context cancelled: %v", ctx.Err())
			}
			return
		case <-st.stop:
			return
		}
	}
}

func (st *Tracker) subscribeOnAddrChanges(ctx context.Context, addrChan chan<- common.Address) {
	st.wg.Add(1)
	defer st.wg.Done()

	events := make(chan *polygonvalidiumetrog.PolygonvalidiumetrogSetTrustedSequencer)
	defer close(events)

	var sub event.Subscription

	initSubscription := func() {
		if err := backoff.Exponential(func() (err error) {
			if sub, err = st.em.WatchSetTrustedSequencer(ctx, events); err != nil {
				log.Errorf("error subscribing to trusted sequencer event, retrying: %v", err)
			}

			return err
		}, maxConnectionRetries, st.retry); err != nil {
			log.Fatalf("failed subscribing to trusted sequencer event: %v. Check ws(s) availability.", err)
		}
	}

	initSubscription()

	for {
		select {
		case e := <-events:
			addrChan <- e.NewTrustedSequencer
		case <-ctx.Done():
			return
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

func (st *Tracker) pollAddrChanges(ctx context.Context, addrChan chan<- common.Address) {
	st.wg.Add(1)
	defer st.wg.Done()

	ticker := time.NewTicker(st.pollInterval)
	for {
		select {
		case <-ticker.C:
			addr, err := st.em.TrustedSequencer(ctx)
			if err != nil {
				log.Errorf("failed to get sequencer addr: %v", err)
				break
			}

			if st.GetAddr().Cmp(addr) != 0 {
				addrChan <- addr
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-st.stop:
			ticker.Stop()
			return
		}
	}
}

func (st *Tracker) trackUrlChanges(ctx context.Context) {
	urlChan := make(chan string, 1)

	if st.usePolling {
		go st.pollUrlChanges(ctx, urlChan)
	} else {
		go st.subscribeOnUrlChanges(ctx, urlChan)
	}

	for {
		select {
		case url := <-urlChan:
			if st.GetUrl() != url {
				log.Infof("new trusted sequencer url: %v", url)
				st.setUrl(url)
			}
		case <-ctx.Done():
			if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
				log.Warnf("context cancelled: %v", ctx.Err())
			}
			return
		case <-st.stop:
			return
		}
	}
}

func (st *Tracker) subscribeOnUrlChanges(ctx context.Context, urlChan chan<- string) {
	st.wg.Add(1)
	defer st.wg.Done()

	events := make(chan *polygonvalidiumetrog.PolygonvalidiumetrogSetTrustedSequencerURL)
	defer close(events)

	var sub event.Subscription

	initSubscription := func() {
		if err := backoff.Exponential(func() (err error) {
			if sub, err = st.em.WatchSetTrustedSequencerURL(ctx, events); err != nil {
				log.Errorf("error subscribing to trusted sequencer URL event, retrying: %v", err)
			}

			return err
		}, maxConnectionRetries, st.retry); err != nil {
			log.Fatalf("failed subscribing to trusted sequencer URL event: %v. Check ws(s) availability.", err)
		}
	}

	initSubscription()

	for {
		select {
		case e := <-events:
			urlChan <- e.NewTrustedSequencerURL
		case <-ctx.Done():
			return
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

func (st *Tracker) pollUrlChanges(ctx context.Context, urlChan chan<- string) {
	st.wg.Add(1)
	defer st.wg.Done()

	ticker := time.NewTicker(st.pollInterval)
	for {
		select {
		case <-ticker.C:
			url, err := st.em.TrustedSequencerURL(ctx)
			if err != nil {
				log.Errorf("failed to get sequencer URL: %v", err)
				break
			}

			if st.GetUrl() != url {
				urlChan <- url
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-st.stop:
			ticker.Stop()
			return
		}
	}
}

// GetSequenceBatch returns sequence batch for given batch number
func (st *Tracker) GetSequenceBatch(ctx context.Context, batchNum uint64) (*SeqBatch, error) {
	return GetData(ctx, st.GetUrl(), batchNum)
}

// Stop stops the SequencerTracker
func (st *Tracker) Stop() {
	close(st.stop)
	st.wg.Wait()
}
