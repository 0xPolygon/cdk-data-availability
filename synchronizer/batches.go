package synchronizer

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/0xPolygon/cdk-data-availability/client"
	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const defaultBlockBatchSize = 32

// BatchSynchronizer watches for batch events, checks if they are "locally" stored, then retrieves and stores missing data
type BatchSynchronizer struct {
	client         *etherman.Etherman
	stop           chan struct{}
	retry          time.Duration
	rpcTimeout     time.Duration
	blockBatchSize uint
	self           common.Address
	db             *db.DB
	committee      map[common.Address]etherman.DataCommitteeMember
	lock           sync.Mutex
	reorgs         <-chan BlockReorg
	events         chan *cdkvalidium.CdkvalidiumSequenceBatches
	sequencer      *sequencer.SequencerTracker
}

// NewBatchSynchronizer creates the BatchSynchronizer
func NewBatchSynchronizer(
	cfg config.L1Config,
	self common.Address,
	db *db.DB,
	reorgs <-chan BlockReorg,
	ethClient *etherman.Etherman,
	sequencer *sequencer.SequencerTracker,
) (*BatchSynchronizer, error) {
	if cfg.BlockBatchSize == 0 {
		log.Infof("block batch size is not set, setting to default %d", defaultBlockBatchSize)
		cfg.BlockBatchSize = defaultBlockBatchSize
	}
	synchronizer := &BatchSynchronizer{
		client:         ethClient,
		stop:           make(chan struct{}),
		retry:          cfg.RetryPeriod.Duration,
		rpcTimeout:     cfg.Timeout.Duration,
		blockBatchSize: cfg.BlockBatchSize,
		self:           self,
		db:             db,
		reorgs:         reorgs,
		events:         make(chan *cdkvalidium.CdkvalidiumSequenceBatches),
		sequencer:      sequencer,
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
func (bs *BatchSynchronizer) Start() {
	log.Infof("starting batch synchronizer, DAC addr: %v", bs.self)
	go bs.consumeEvents()
	go bs.produceEvents()
	go bs.handleReorgs()
}

// Stop stops the synchronizer
func (bs *BatchSynchronizer) Stop() {
	close(bs.events)
	close(bs.stop)
}

func (bs *BatchSynchronizer) handleReorgs() {
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
		case <-bs.stop:
			return
		}
	}
}

func (bs *BatchSynchronizer) produceEvents() {
	log.Info("starting event producer")
	for {
		delay := time.NewTimer(bs.retry)
		select {
		case <-delay.C:
			if err := bs.filterEvents(); err != nil {
				log.Errorf("error filtering events: %v", err)
			}
		case <-bs.stop:
			return
		}
	}
}

// Start an iterator from last block processed, picking off SequenceBatches events
func (bs *BatchSynchronizer) filterEvents() error {
	start, err := getStartBlock(bs.db)
	if err != nil {
		return err
	}

	end := start + uint64(bs.blockBatchSize)

	// get the latest block number
	header, err := bs.client.EthClient.HeaderByNumber(context.TODO(), nil)
	if err != nil {
		log.Errorf("failed to determine latest block number", err)
		return err
	}
	// we don't want to scan beyond latest block
	if end > header.Number.Uint64() {
		end = header.Number.Uint64()
	}

	iter, err := bs.client.CDKValidium.FilterSequenceBatches(
		&bind.FilterOpts{
			Start:   start,
			End:     &end,
			Context: context.TODO(),
		}, nil)
	if err != nil {
		return err
	}
	for iter.Next() {
		if iter.Error() != nil {
			return iter.Error()
		}
		bs.events <- iter.Event
	}

	// advance start block
	err = setStartBlock(bs.db, end)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BatchSynchronizer) consumeEvents() {
	log.Info("starting event consumer")
	for {
		select {
		case sb := <-bs.events:
			if err := bs.handleEvent(sb); err != nil {
				log.Errorf("failed to handle event: %v", err)
			}
		case <-bs.stop:
			return
		}
	}
}

func (bs *BatchSynchronizer) handleEvent(event *cdkvalidium.CdkvalidiumSequenceBatches) error {
	ctx, cancel := context.WithTimeout(context.Background(), bs.rpcTimeout)
	defer cancel()

	tx, _, err := bs.client.GetTx(ctx, event.Raw.TxHash)
	if err != nil {
		return err
	}
	txData := tx.Data()
	_, keys, err := etherman.ParseEvent(event, txData)
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
	if len(missing) == 0 {
		return nil
	}

	var data []types.OffChainData
	for _, key := range missing {
		log.Infof("resolving missing key %v", key.Hex())
		var value *types.OffChainData
		value, err = bs.resolve(event.NumBatch, key)
		if err != nil {
			return err
		}
		data = append(data, *value)
	}

	// Finally, store the data
	return store(bs.db, data)
}

func (bs *BatchSynchronizer) resolve(batchNum uint64, key common.Hash) (*types.OffChainData, error) {
	log.Debugf("resolving missing data for key %v", key.Hex())

	data := bs.trySequencer(batchNum, key)
	if data != nil {
		return data, nil
	}

	// sequencer failed to produce data, try the other nodes
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
		log.Infof("trying DAC %s: %s", member.Addr.Hex(), member.URL)
		value, err := bs.resolveWithMember(key, member)
		if err != nil {
			log.Warnf("error resolving, continuing: %v", err)
			delete(bs.committee, member.Addr)
			continue // did not have data or errored out
		}

		return value, nil
	}
	return nil, rpc.NewRPCError(rpc.NotFoundErrorCode, "no data found for key %v", key)
}

// trySequencer returns L2Data from the trusted sequencer, but does not return errors, only logs warnings if not found.
func (bs *BatchSynchronizer) trySequencer(batchNum uint64, key common.Hash) *types.OffChainData {
	log.Debugf("resolving batch %d, key %s, with sequencer at %s", batchNum, key.Hex(), bs.sequencer.GetUrl())
	data, err := sequencer.GetData(bs.sequencer.GetUrl(), batchNum)
	if err != nil {
		log.Warnf("failed to get data from sequencer: %v", err)
		return nil
	}
	expectKey := crypto.Keccak256Hash(data.L2Data)
	if key != expectKey {
		log.Warnf("sequencer gave wrong data for key: %s", key.Hex())
		return nil
	}
	return &types.OffChainData{
		Key:   key,
		Value: data.L2Data,
	}
}

func (bs *BatchSynchronizer) resolveWithMember(key common.Hash, member etherman.DataCommitteeMember) (*types.OffChainData, error) {
	cm := client.New(member.URL)
	ctx, cancel := context.WithTimeout(context.Background(), bs.rpcTimeout)
	defer cancel()

	log.Debugf("trying member %v at %v for key %v", member.Addr.Hex(), member.URL, key.Hex())

	bytes, err := cm.GetOffChainData(ctx, key)
	if err != nil {
		return nil, err
	}
	expectKey := crypto.Keccak256Hash(bytes)
	if key != expectKey {
		return nil, err
	}
	return &types.OffChainData{
		Key:   key,
		Value: bytes,
	}, nil
}
