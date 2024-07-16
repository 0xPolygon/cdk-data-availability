package synchronizer

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonvalidiumetrog"
	"github.com/0xPolygon/cdk-data-availability/client"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/etherman"
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
	GetSequenceBatch(ctx context.Context, batchNum uint64) (*sequencer.SeqBatch, error)
}

// BatchSynchronizer watches for number events, checks if they are
// "locally" stored, then retrieves and stores missing data
type BatchSynchronizer struct {
	client               etherman.Etherman
	stop                 chan struct{}
	retry                time.Duration
	rpcTimeout           time.Duration
	blockBatchSize       uint
	self                 common.Address
	db                   db.DB
	committee            *CommitteeMapSafe
	syncLock             sync.Mutex
	reorgs               <-chan BlockReorg
	events               chan *polygonvalidiumetrog.PolygonvalidiumetrogSequenceBatches
	sequencer            SequencerTracker
	rpcClientFactory     client.Factory
	offchainDataGaps     map[uint64]uint64
	offchainDataGapsLock sync.RWMutex
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
		events:           make(chan *polygonvalidiumetrog.PolygonvalidiumetrogSequenceBatches),
		sequencer:        sequencer,
		rpcClientFactory: rpcClientFactory,
		offchainDataGaps: make(map[uint64]uint64),
	}
	return synchronizer, synchronizer.resolveCommittee()
}

func (bs *BatchSynchronizer) resolveCommittee() error {
	current, err := bs.client.GetCurrentDataCommittee()
	if err != nil {
		return err
	}

	filteredMembers := make([]etherman.DataCommitteeMember, 0, len(current.Members))
	for _, m := range current.Members {
		if m.Addr != bs.self {
			filteredMembers = append(filteredMembers, m)
		}
	}

	bs.committee = NewCommitteeMapSafe()
	bs.committee.StoreBatch(filteredMembers)
	return nil
}

// Start starts the synchronizer
func (bs *BatchSynchronizer) Start(ctx context.Context) {
	log.Infof("starting batch synchronizer, DAC addr: %v", bs.self)
	go bs.processUnresolvedBatches(ctx)
	go bs.produceEvents(ctx)
	go bs.handleReorgs(ctx)
	go bs.startOffchainDataGapsDetection(ctx)
}

// Stop stops the synchronizer
func (bs *BatchSynchronizer) Stop() {
	close(bs.stop)
}

// Gaps returns the offchain data gaps
func (bs *BatchSynchronizer) Gaps() map[uint64]uint64 {
	bs.offchainDataGapsLock.RLock()
	gaps := make(map[uint64]uint64, len(bs.offchainDataGaps))
	for key, value := range bs.offchainDataGaps {
		gaps[key] = value
	}
	bs.offchainDataGapsLock.RUnlock()
	return gaps
}

func (bs *BatchSynchronizer) handleReorgs(ctx context.Context) {
	log.Info("starting reorgs handler")
	for {
		select {
		case r := <-bs.reorgs:
			bs.syncLock.Lock()

			latest, err := getStartBlock(ctx, bs.db, L1SyncTask)
			if err != nil {
				log.Errorf("could not determine latest processed block: %v", err)
				bs.syncLock.Unlock()

				continue
			}

			if latest < r.Number {
				// only reset start block if necessary
				bs.syncLock.Unlock()
				continue
			}

			if err = setStartBlock(ctx, bs.db, r.Number, L1SyncTask); err != nil {
				log.Errorf("failed to store new start block to %d: %v", r.Number, err)
			}

			bs.syncLock.Unlock()
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
	bs.syncLock.Lock()
	defer bs.syncLock.Unlock()

	start, err := getStartBlock(ctx, bs.db, L1SyncTask)
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
	var events []*polygonvalidiumetrog.PolygonvalidiumetrogSequenceBatches
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
			return setStartBlock(ctx, bs.db, event.Raw.BlockNumber-1, L1SyncTask)
		}
	}

	return setStartBlock(ctx, bs.db, end, L1SyncTask)
}

func (bs *BatchSynchronizer) handleEvent(
	parentCtx context.Context,
	event *polygonvalidiumetrog.PolygonvalidiumetrogSequenceBatches,
) error {
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

func (bs *BatchSynchronizer) processUnresolvedBatches(ctx context.Context) {
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

	// Collect list of keys
	keys := make([]common.Hash, len(batchKeys))
	hashToKeys := make(map[common.Hash]types.BatchKey)
	for i, key := range batchKeys {
		keys[i] = key.Hash
		hashToKeys[key.Hash] = key
	}

	// Get the existing offchain data by the given list of keys
	existingOffchainData, err := listOffchainData(ctx, bs.db, keys)
	if err != nil {
		return fmt.Errorf("failed to list offchain data: %v", err)
	}

	// Resolve the unresolved data
	data := make([]types.OffChainData, 0)
	resolved := make([]types.BatchKey, 0)

	// Go over existing keys and mark them as resolved if they exist.
	// Update the batch number if it is zero.
	for _, extData := range existingOffchainData {
		batchKey, ok := hashToKeys[extData.Key]
		if !ok {
			// This should not happen, but log it just in case
			log.Errorf("unexpected key %s in the offchain data", extData.Key.Hex())
			continue
		}

		// If the batch number is zero, update it
		if extData.BatchNum == 0 {
			extData.BatchNum = batchKey.Number
			data = append(data, extData)
		}

		// Mark the batch as resolved
		resolved = append(resolved, batchKey)

		// Remove the key from the map
		delete(hashToKeys, extData.Key)
	}

	// Resolve the remaining unresolved data
	for _, key := range hashToKeys {
		value, err := bs.resolve(ctx, key)
		if err != nil {
			log.Errorf("failed to resolve batch %s: %v", key.Hash.Hex(), err)
			continue
		}

		resolved = append(resolved, key)
		data = append(data, *value)
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

func (bs *BatchSynchronizer) resolve(ctx context.Context, batch types.BatchKey) (*types.OffChainData, error) {
	// First try to get the data from the trusted sequencer
	data := bs.trySequencer(ctx, batch)
	if data != nil {
		return data, nil
	}

	// If the sequencer failed to produce data, try the other nodes
	if bs.committee.Length() == 0 {
		// committee is resolved again once all members are evicted. They can be evicted
		// for not having data, or their config being malformed
		if err := bs.resolveCommittee(); err != nil {
			return nil, err
		}
	}

	// pull out the members, iterating will change the map on error
	members := bs.committee.AsSlice()

	// iterate through them randomly until data is resolved
	for _, r := range rand.Perm(len(members)) {
		member := members[r]
		if member.URL == "" ||
			common.HexToAddress("0x0").Cmp(member.Addr) == 0 ||
			member.Addr.Cmp(bs.self) == 0 {
			bs.committee.Delete(member.Addr)
			continue // malformed committee, skip what is known to be wrong
		}

		value, err := bs.resolveWithMember(ctx, batch, member)
		if err != nil {
			log.Warnf("error resolving, continuing: %v", err)
			bs.committee.Delete(member.Addr)
			continue // did not have data or errored out
		}

		return value, nil
	}

	return nil, rpc.NewRPCError(rpc.NotFoundErrorCode,
		"no data found for number %d, key %v", batch.Number, batch.Hash.Hex())
}

// trySequencer returns L2Data from the trusted sequencer, but does not return errors, only logs warnings if not found.
func (bs *BatchSynchronizer) trySequencer(ctx context.Context, batch types.BatchKey) *types.OffChainData {
	seqBatch, err := bs.sequencer.GetSequenceBatch(ctx, batch.Number)
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
		Key:      batch.Hash,
		Value:    seqBatch.BatchL2Data,
		BatchNum: batch.Number,
	}
}

func (bs *BatchSynchronizer) resolveWithMember(
	parentCtx context.Context,
	batch types.BatchKey,
	member etherman.DataCommitteeMember,
) (*types.OffChainData, error) {
	cm := bs.rpcClientFactory.New(member.URL)

	ctx, cancel := context.WithTimeout(parentCtx, bs.rpcTimeout)
	defer cancel()

	log.Debugf("trying member %v at %v for key %v", member.Addr.Hex(), member.URL, batch.Hash.Hex())

	bytes, err := cm.GetOffChainData(ctx, batch.Hash)
	if err != nil {
		return nil, err
	}

	expectKey := crypto.Keccak256Hash(bytes)
	if batch.Hash.Cmp(expectKey) != 0 {
		return nil, fmt.Errorf("unexpected key gotten from member: %v. Key: %v", member.Addr.Hex(), expectKey.Hex())
	}

	return &types.OffChainData{
		Key:      batch.Hash,
		Value:    bytes,
		BatchNum: batch.Number,
	}, nil
}

func (bs *BatchSynchronizer) startOffchainDataGapsDetection(ctx context.Context) {
	log.Info("starting handling unresolved batches")
	for {
		delay := time.NewTimer(time.Minute)
		select {
		case <-delay.C:
			if err := bs.detectOffchainDataGaps(ctx); err != nil {
				log.Error(err)
			}
		case <-bs.stop:
			return
		}
	}
}

// detectOffchainDataGaps detects offchain data gaps and reports them in logs and the service state.
func (bs *BatchSynchronizer) detectOffchainDataGaps(ctx context.Context) error {
	// Detect offchain data gaps
	gaps, err := detectOffchainDataGaps(ctx, bs.db)
	if err != nil {
		return fmt.Errorf("failed to detect offchain data gaps: %v", err)
	}

	// No gaps found, all good
	if len(gaps) == 0 {
		return nil
	}

	// Log the detected gaps and store the detected gaps in the service state
	gapsRaw := new(bytes.Buffer)
	bs.offchainDataGapsLock.Lock()
	bs.offchainDataGaps = make(map[uint64]uint64, len(gaps))
	for key, value := range gaps {
		bs.offchainDataGaps[key] = value

		if _, err = fmt.Fprintf(gapsRaw, "%d=>%d\n", key, value); err != nil {
			log.Errorf("failed to write offchain data gaps: %v", err)
		}
	}
	bs.offchainDataGapsLock.Unlock()

	log.Warnf("detected offchain data gaps (current batch number => expected batch number): %s", gapsRaw.String())

	return nil
}
