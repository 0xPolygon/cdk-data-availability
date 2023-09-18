package synchronizer

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/offchaindata"
	"github.com/0xPolygon/cdk-validium-node/etherman"
	"github.com/0xPolygon/cdk-validium-node/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-validium-node/jsonrpc/types"
	"github.com/0xPolygon/cdk-validium-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// BatchSynchronizer watches for batch events, checks if they are "locally" stored, then retrieves and stores missing data
type BatchSynchronizer struct {
	client    *etherman.Client
	stop      chan struct{}
	timeout   time.Duration
	retry     time.Duration
	self      common.Address
	db        *db.DB
	committee map[common.Address]etherman.DataCommitteeMember
	lock      sync.Mutex
	reorgs    <-chan BlockReorg
}

// NewBatchSynchronizer creates the BatchSynchronizer
func NewBatchSynchronizer(cfg config.L1Config, self common.Address, db *db.DB, reorgs <-chan BlockReorg) (*BatchSynchronizer, error) {
	ethClient, err := newEtherman(cfg)
	if err != nil {
		return nil, err
	}
	synchronizer := &BatchSynchronizer{
		client:  ethClient,
		stop:    make(chan struct{}),
		timeout: cfg.Timeout.Duration,
		retry:   cfg.RetryPeriod.Duration,
		self:    self,
		db:      db,
		reorgs:  reorgs,
	}
	err = synchronizer.resolveCommittee()
	if err != nil {
		return nil, err
	}
	return synchronizer, nil
}

func (bs *BatchSynchronizer) resolveCommittee() error {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	committee := make(map[common.Address]etherman.DataCommitteeMember)
	current, err := bs.client.GetCurrentDataCommittee()
	if err != nil {
		return err
	}
	for _, member := range current.Members {
		if bs.self != member.Addr {
			committee[member.Addr] = member
		}
	}
	bs.committee = committee
	return nil
}

func (bs *BatchSynchronizer) Start() {
	log.Info("starting batch synchronizer")

	events := make(chan *cdkvalidium.CdkvalidiumSequenceBatches)
	defer close(events)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go bs.consumeEvents(ctx, events)
	go bs.produceEvents(ctx, events)
	go bs.handleReorgs(ctx)

	<-bs.stop
}

func (bs *BatchSynchronizer) Stop() {
	close(bs.stop)
}

func (bs *BatchSynchronizer) handleReorgs(ctx context.Context) {
	for {
		select {
		case r := <-bs.reorgs:
			latest, err := getStartBlock(bs.db)
			if err != nil {
				log.Errorf("could not determine latest processed block: %v", err)
				continue
			}
			if latest < r.Number {
				// only reset start block if necessary
				continue
			}
			err = setStartBlock(bs.db, r.Number)
			if err != nil {
				log.Errorf("failed to store new start block to %d: %v", r.Number, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (bs *BatchSynchronizer) produceEvents(ctx context.Context, events chan *cdkvalidium.CdkvalidiumSequenceBatches) {
	for {
		delay := time.NewTimer(bs.retry)
		select {
		case <-delay.C:
			if err := bs.filterEvents(ctx, events); err != nil {
				log.Errorf("error filtering events: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Start an iterator from last block processed, picking off SequenceBatches events
func (bs *BatchSynchronizer) filterEvents(ctx context.Context, events chan *cdkvalidium.CdkvalidiumSequenceBatches) error {
	start, err := getStartBlock(bs.db)
	if err != nil {
		return err
	}
	iter, err := bs.client.CDKValidium.FilterSequenceBatches(
		&bind.FilterOpts{
			Start:   start,
			Context: ctx,
		}, nil)
	if err != nil {
		return err
	}
	for iter.Next() {
		// NOTE: batch number is _not_ block number
		log.Debugf("filter event batch number %d in block %d", iter.Event.NumBatch, iter.Event.Raw.BlockNumber)
		events <- iter.Event
	}
	return nil
}

func (bs *BatchSynchronizer) consumeEvents(ctx context.Context, events chan *cdkvalidium.CdkvalidiumSequenceBatches) {
	for {
		select {
		case sb := <-events:
			if err := bs.handleEvent(sb); err != nil {
				log.Errorf("failed to handle event: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (bs *BatchSynchronizer) handleEvent(event *cdkvalidium.CdkvalidiumSequenceBatches) error {
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()

	tx, _, err := bs.client.GetTx(ctx, event.Raw.TxHash)
	if err != nil {
		return err
	}
	txData := tx.Data()
	block, keys, err := parseEvent(event, txData)
	if err != nil {
		return err
	}

	// collect keys that need to be resolved
	var missing []common.Hash
	for _, key := range keys {
		if !exists(bs.db, key) { // this could be a single query that takes the whole list and returns missing ones
			missing = append(missing, key)
		}
	}

	log.Debugf("checking data %d missing keys", len(missing))

	var data []offchaindata.OffChainData
	for _, key := range missing {
		var value offchaindata.OffChainData
		value, err = bs.resolve(key)
		if err != nil {
			return err // return so that the block does not get updated in sync info
		}
		data = append(data, value)
	}

	// Finally, store the data
	return store(bs.db, block, data)
}

func (bs *BatchSynchronizer) resolve(key common.Hash) (offchaindata.OffChainData, error) {
	log.Debugf("resolving missing data for key %v", key.Hex())
	if len(bs.committee) == 0 {
		err := bs.resolveCommittee()
		if err != nil {
			return offchaindata.OffChainData{}, err
		}
	}
	// pull out the members, iterating will change the map on error
	members := make([]etherman.DataCommitteeMember, len(bs.committee))
	for _, member := range bs.committee {
		members = append(members, member)
	}
	// iterate through them randomly until data is resolved
	for _, r := range rand.Perm(len(members)) {
		member := members[r]
		value, err := resolveWithMember(key, member)
		if err != nil {
			log.Warnf("resolve member %v failed, removing from local committee cache: %v", member.Addr, err)
			delete(bs.committee, member.Addr)
			continue // did not have data or errored out
		}
		return value, nil
	}
	return offchaindata.OffChainData{}, types.NewRPCError(types.NotFoundErrorCode, "no data found for key %v", key)
}
