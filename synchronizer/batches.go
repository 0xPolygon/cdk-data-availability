package synchronizer

import (
	"context"
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/0xPolygon/supernets2-data-availability/client"
	"github.com/0xPolygon/supernets2-data-availability/config"
	"github.com/0xPolygon/supernets2-data-availability/db"
	"github.com/0xPolygon/supernets2-data-availability/offchaindata"
	"github.com/0xPolygon/supernets2-node/etherman"
	"github.com/0xPolygon/supernets2-node/etherman/smartcontracts/supernets2"
	"github.com/0xPolygon/supernets2-node/jsonrpc/types"
	"github.com/0xPolygon/supernets2-node/log"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/jackc/pgx/v4"
)

// BatchSynchronizer watches for batch events, checks if they are "locally" stored, then retrieves and stores missing data
type BatchSynchronizer struct {
	watcher
	db        *db.DB
	committee *DataCommitteeTracker
}

// NewBatchSynchronizer creates the BatchSynchronizer
func NewBatchSynchronizer(cfg config.L1Config, committee *DataCommitteeTracker) (*BatchSynchronizer, error) {
	watcher, err := newWatcher(cfg)
	if err != nil {
		return nil, err
	}
	return &BatchSynchronizer{
		watcher:   *watcher,
		committee: committee,
	}, nil
}

// Start starts the BatchSynchronizer event subscription
func (bs *BatchSynchronizer) Start() {
	events := make(chan *supernets2.Supernets2SequenceBatches)
	defer close(events)
	for {
		var (
			sub   event.Subscription
			err   error
			start uint64
		)

		start, err = bs.getStartBlock()

		ctx, cancel := context.WithTimeout(context.Background(), bs.timeout)
		opts := &bind.WatchOpts{Context: ctx, Start: &start}
		sub, err = bs.client.Supernets2.WatchSequenceBatches(opts, events, nil)

		// if no subscription, retry until established
		for err != nil {
			<-time.After(bs.retry)
			sub, err = bs.client.Supernets2.WatchSequenceBatches(opts, events, nil)
			if err != nil {
				log.Errorf("error subscribing to sequence batch events, retrying: %v", err)
			}
		}

		// wait on events, timeouts, and signals to stop
		select {
		case sb := <-events:
			err = bs.handleSequenceBatches(sb)
			if err != nil {
				log.Errorf("failed to process batches: %v", sb)
				sub.Unsubscribe()
				continue // restart subscription
			}
		case err := <-sub.Err():
			log.Warnf("subscription error, resubscribing: %v", err)
		case <-ctx.Done():
			if ctx.Err() != nil {
				log.Warn("re-establishing subscription: %v", ctx.Err())
			}
		case <-bs.stop:
			if sub != nil {
				sub.Unsubscribe()
			}
			cancel()
			return
		}
	}
}

func (bs *BatchSynchronizer) getStartBlock() (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // TODO: make configurable
	defer cancel()

	start, err := bs.db.GetLastProcessedBlock(ctx)
	if err != nil {
		log.Errorf("error retrieving last processed block, starting from 0: %v", err)
	}
	if start > 0 {
		start = start - 1 // since a block may have been partially processed
	}
	return start, err
}

// Stop stops the BatchSynchronizer
func (bs *BatchSynchronizer) Stop() {
	close(bs.stop)
}

func (bs *BatchSynchronizer) handleSequenceBatches(event *supernets2.Supernets2SequenceBatches) error {
	block, keys, err := parseEvent(event)
	if err != nil {
		return err
	}

	// collect keys that need to be resolved
	var missing []common.Hash
	for _, key := range keys {
		if !bs.exists(key) {
			missing = append(missing, key)
		}
	}
	return bs.resolveAndStore(block, missing)
}

func (bs *BatchSynchronizer) exists(key common.Hash) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second) // TODO: make configurable
	defer cancel()
	return bs.db.Exists(ctx, key)
}

func (bs *BatchSynchronizer) resolveAndStore(block uint64, keys []common.Hash) error {
	var data []offchaindata.OffChainData
	for _, key := range keys {
		value, err := bs.resolve(key)
		if err != nil {
			return err // return so that the block does not get updated in sync info
		}
		data = append(data, value)
	}
	return bs.store(block, data)
}

func (bs *BatchSynchronizer) store(block uint64, data []offchaindata.OffChainData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // TODO: configure
	defer cancel()
	var (
		dbTx pgx.Tx
		err  error
	)
	if dbTx, err = bs.db.BeginStateTransaction(ctx); err != nil {
		return err
	}
	if err = bs.db.StoreOffChainData(ctx, data, dbTx); err != nil {
		rollback(ctx, err, dbTx)
		return err
	}
	if err = bs.db.StoreLastProcessedBlock(ctx, block, dbTx); err != nil {
		rollback(ctx, err, dbTx)
		return err
	}
	if err = dbTx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func rollback(ctx context.Context, err error, dbTx pgx.Tx) {
	if txErr := dbTx.Rollback(ctx); txErr != nil {
		log.Errorf("failed to roll back transaction after error %v : %v", err, txErr)
	}
}

func (bs *BatchSynchronizer) resolve(key common.Hash) (offchaindata.OffChainData, error) {
	rand.NewSource(time.Now().UnixNano())
	committee := bs.committee.GetMembers().Members
	random := rand.Perm(len(committee))
	for i := 0; i < len(random); i++ {
		value, err := resolveWithMember(key, committee[i])
		if err != nil {
			continue // did not have data or errored out
		}
		return value, nil

	}
	return offchaindata.OffChainData{}, types.NewRPCError(types.NotFoundErrorCode, "no data found for key %v", key)
}

func resolveWithMember(key common.Hash, member etherman.DataCommitteeMember) (offchaindata.OffChainData, error) {
	cm := client.New(member.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // TODO: configure
	defer cancel()
	bytes, err := cm.GetOffChainData(ctx, key)
	if len(bytes) == 0 {
		err = types.NewRPCError(types.NotFoundErrorCode, "data not found")
	}
	var data offchaindata.OffChainData
	if len(bytes) > 0 {
		data = offchaindata.OffChainData{
			Key:   key,
			Value: bytes,
		}
	}
	return data, err
}

func parseEvent(event *supernets2.Supernets2SequenceBatches) (uint64, []common.Hash, error) {
	a, err := abi.JSON(strings.NewReader(supernets2.Supernets2ABI))
	if err != nil {
		return 0, nil, err
	}
	method, err := a.MethodById(event.Raw.Data[:4])
	if err != nil {
		return 0, nil, err
	}
	data, err := method.Inputs.Unpack(event.Raw.Data[4:])
	if err != nil {
		return 0, nil, err
	}
	var batches []supernets2.Supernets2BatchData
	bytes, err := json.Marshal(data[0])
	if err != nil {
		return 0, nil, err
	}
	err = json.Unmarshal(bytes, &batches)
	if err != nil {
		return 0, nil, err
	}

	var keys []common.Hash
	for _, batch := range batches {
		keys = append(keys, batch.TransactionsHash)
	}
	return event.Raw.BlockNumber, keys, nil
}
