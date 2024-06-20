package synchronizer

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/client"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/etrog/polygonvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const defaultBlockBatchSize = 32

// SequencerTracker is an interface that defines functions that a sequencer tracker must implement
type SequencerTracker interface {
	GetSequenceBatch(batchNum uint64) (*sequencer.SeqBatch, error)
}

// BatchSynchronizer watches for number events, checks if they are "locally" stored, then retrieves and stores missing data
type BatchSynchronizer struct {
	client           etherman.Etherman
	stop             chan struct{}
	retry            time.Duration
	rpcTimeout       time.Duration
	blockBatchSize   uint
	self             common.Address
	db               db.DB
	committee        map[common.Address]etherman.DataCommitteeMember
	lock             sync.Mutex
	reorgs           <-chan BlockReorg
	sequencer        SequencerTracker
	rpcClientFactory client.Factory
}

// NewBatchSynchronizer creates the BatchSynchronizer
func NewBatchSynchronizer(
	cfg config.L1Config,
	self common.Address,
	db db.DB,
	reorgs <-chan BlockReorg,
	ethClient etherman.Etherman,
	sequencer SequencerTracker,
	rpcClientFactory client.Factory,
) (*BatchSynchronizer, error) {
	if cfg.BlockBatchSize == 0 {
		log.Infof("block number size is not set, setting to default %d", defaultBlockBatchSize)
		cfg.BlockBatchSize = defaultBlockBatchSize
	}
	synchronizer := &BatchSynchronizer{
		client:           ethClient,
		stop:             make(chan struct{}),
		retry:            cfg.RetryPeriod.Duration,
		rpcTimeout:       cfg.Timeout.Duration,
		blockBatchSize:   cfg.BlockBatchSize,
		self:             self,
		db:               db,
		reorgs:           reorgs,
		sequencer:        sequencer,
		rpcClientFactory: rpcClientFactory,
	}
	return synchronizer, synchronizer.resolveCommittee()
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

// Start starts the synchronizer
func (bs *BatchSynchronizer) Start(ctx context.Context) {
	log.Infof("starting batch synchronizer, DAC addr: %v", bs.self)
	go bs.startUnresolvedBatchesProcessor(ctx)
	go bs.produceEvents(ctx)
	go bs.handleReorgs(ctx)
}

// Stop stops the synchronizer
func (bs *BatchSynchronizer) Stop() {
	close(bs.stop)
}

func (bs *BatchSynchronizer) handleReorgs(ctx context.Context) {
	log.Info("starting reorgs handler")
	for {
		select {
		case r := <-bs.reorgs:
			latest, err := getStartBlock(ctx, bs.db)
			if err != nil {
				log.Errorf("could not determine latest processed block: %v", err)
				continue
			}

			if latest < r.Number {
				// only reset start block if necessary
				continue
			}

			if err = setStartBlock(ctx, bs.db, r.Number); err != nil {
				log.Errorf("failed to store new start block to %d: %v", r.Number, err)
			}
		case <-bs.stop:
			return
		}
	}
}

func (bs *BatchSynchronizer) produceEvents(ctx context.Context) {
	log.Info("starting event producer")
	for {
		delay := time.NewTimer(bs.retry)
		select {
		case <-delay.C:
			if err := bs.filterEvents(ctx); err != nil {
				log.Errorf("error filtering events: %v", err)
			}
		case <-bs.stop:
			return
		}
	}
}

// Start an iterator from last block processed, picking off SequenceBatches events
func (bs *BatchSynchronizer) filterEvents(ctx context.Context) error {
	start, err := getStartBlock(ctx, bs.db)
	if err != nil {
		return err
	}

	end := start + uint64(bs.blockBatchSize)

	// get the latest block number
	header, err := bs.client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Errorf("failed to determine latest block number: %v", err)
		return err
	}

	// we don't want to scan beyond latest block
	if end > header.Number.Uint64() {
		end = header.Number.Uint64()
	}

	iter, err := bs.client.FilterSequenceBatches(
		&bind.FilterOpts{
			Context: ctx,
			Start:   start,
			End:     &end,
		}, nil)
	if err != nil {
		log.Errorf("failed to create SequenceBatches event iterator: %v", err)
		return err
	}

	// Collect events into the slice
	var events []*polygonvalidium.PolygonvalidiumSequenceBatches
	for iter.Next() {
		if iter.Error() != nil {
			return iter.Error()
		}

		events = append(events, iter.Event)
	}

	if err = iter.Close(); err != nil {
		log.Errorf("failed to close SequenceBatches event iterator: %v", err)
	}

	// Sort events by block number ascending
	sort.Slice(events, func(i, j int) bool {
		return events[i].Raw.BlockNumber < events[j].Raw.BlockNumber
	})

	// Handle events
	for _, event := range events {
		if err = bs.handleEvent(ctx, event); err != nil {
			log.Errorf("failed to handle event: %v", err)
			return setStartBlock(ctx, bs.db, event.Raw.BlockNumber-1)
		}
	}

	return setStartBlock(ctx, bs.db, end)
}

func (bs *BatchSynchronizer) handleEvent(parentCtx context.Context, event *polygonvalidium.PolygonvalidiumSequenceBatches) error {
	ctx, cancel := context.WithTimeout(parentCtx, bs.rpcTimeout)
	defer cancel()

	tx, _, err := bs.client.GetTx(ctx, event.Raw.TxHash)
	if err != nil {
		return err
	}

	keys, err := UnpackTxData(tx.Data())
	if err != nil {
		return err
	}

	// The event has the _last_ batch number & list of hashes. Each hash is
	// in order, so the batch number can be computed from position in array
	var batchKeys []types.BatchKey
	for i, j := 0, len(keys)-1; i < len(keys); i, j = i+1, j-1 {
		batchKeys = append(batchKeys, types.BatchKey{
			Number: event.NumBatch - uint64(i),
			Hash:   keys[j],
		})
	}

	// Store batch keys. Already handled batch keys are going to be ignored based on the DB logic.
	return storeUnresolvedBatchKeys(ctx, bs.db, batchKeys)
}

func (bs *BatchSynchronizer) startUnresolvedBatchesProcessor(ctx context.Context) {
	log.Info("starting handling unresolved batches")
	for {
		delay := time.NewTimer(bs.retry)
		select {
		case <-delay.C:
			if err := bs.handleUnresolvedBatches(ctx); err != nil {
				log.Error(err)
			}
		case <-bs.stop:
			return
		}
	}
}

// handleUnresolvedBatches handles unresolved batches that were collected by the event consumer
func (bs *BatchSynchronizer) handleUnresolvedBatches(ctx context.Context) error {
	// Get unresolved batches
	batchKeys, err := getUnresolvedBatchKeys(ctx, bs.db)
	if err != nil {
		return fmt.Errorf("failed to get unresolved batch keys: %v", err)
	}

	if len(batchKeys) == 0 {
		return nil
	}

	// Resolve the unresolved data
	var data []types.OffChainData
	var resolved []types.BatchKey
	for _, key := range batchKeys {
		if exists(ctx, bs.db, key.Hash) {
			resolved = append(resolved, key)
		} else {
			var value *types.OffChainData
			if value, err = bs.resolve(key); err != nil {
				log.Errorf("failed to resolve batch %s: %v", key.Hash.Hex(), err)
				continue
			}

			resolved = append(resolved, key)
			data = append(data, *value)
		}
	}

	// Store data of the batches to the DB
	if len(data) > 0 {
		if err = storeOffchainData(ctx, bs.db, data); err != nil {
			return fmt.Errorf("failed to store offchain data: %v", err)
		}
	}

	// Mark batches as resolved
	if len(resolved) > 0 {
		if err = deleteUnresolvedBatchKeys(ctx, bs.db, resolved); err != nil {
			return fmt.Errorf("failed to delete successfully resolved batch keys: %v", err)
		}
	}

	return nil
}

func (bs *BatchSynchronizer) resolve(batch types.BatchKey) (*types.OffChainData, error) {
	// First try to get the data from the trusted sequencer
	data := bs.trySequencer(batch)
	if data != nil {
		return data, nil
	}

	// If the sequencer failed to produce data, try the other nodes
	if len(bs.committee) == 0 {
		// committee is resolved again once all members are evicted. They can be evicted
		// for not having data, or their config being malformed
		err := bs.resolveCommittee()
		if err != nil {
			return nil, err
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
		if member.URL == "" || member.Addr == common.HexToAddress("0x0") || member.Addr == bs.self {
			delete(bs.committee, member.Addr)
			continue // malformed committee, skip what is known to be wrong
		}
		value, err := bs.resolveWithMember(batch.Hash, member)
		if err != nil {
			log.Warnf("error resolving, continuing: %v", err)
			delete(bs.committee, member.Addr)
			continue // did not have data or errored out
		}

		return value, nil
	}
	return nil, rpc.NewRPCError(rpc.NotFoundErrorCode, "no data found for number %d, key %v", batch.Number, batch.Hash.Hex())
}

// trySequencer returns L2Data from the trusted sequencer, but does not return errors, only logs warnings if not found.
func (bs *BatchSynchronizer) trySequencer(batch types.BatchKey) *types.OffChainData {
	seqBatch, err := bs.sequencer.GetSequenceBatch(batch.Number)
	if err != nil {
		log.Warnf("failed to get data from sequencer: %v", err)
		return nil
	}

	expectKey := crypto.Keccak256Hash(seqBatch.BatchL2Data)
	if batch.Hash != expectKey {
		log.Warnf("number %d: sequencer gave wrong data for key: %s", batch.Number, batch.Hash.Hex())
		return nil
	}
	return &types.OffChainData{
		Key:   batch.Hash,
		Value: seqBatch.BatchL2Data,
	}
}

func (bs *BatchSynchronizer) resolveWithMember(key common.Hash, member etherman.DataCommitteeMember) (*types.OffChainData, error) {
	cm := bs.rpcClientFactory.New(member.URL)
	ctx, cancel := context.WithTimeout(context.Background(), bs.rpcTimeout)
	defer cancel()

	log.Debugf("trying member %v at %v for key %v", member.Addr.Hex(), member.URL, key.Hex())

	bytes, err := cm.GetOffChainData(ctx, key)
	if err != nil {
		return nil, err
	}
	expectKey := crypto.Keccak256Hash(bytes)
	if key != expectKey {
		return nil, fmt.Errorf("unexpected key gotten from member: %v. Key: %v", member.Addr.Hex(), expectKey.Hex())
	}
	return &types.OffChainData{
		Key:   key,
		Value: bytes,
	}, nil
}
