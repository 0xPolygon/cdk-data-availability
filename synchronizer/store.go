package synchronizer

import (
	"context"
	"time"

	dbTypes "github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// SyncTask is the type of the sync task
type SyncTask string

const (
	// L1SyncTask is the name of the L1 sync task
	L1SyncTask SyncTask = "L1"

	dbTimeout = 2 * time.Second
)

func getStartBlock(parentCtx context.Context, db dbTypes.DB, syncTask SyncTask) (uint64, error) {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	start, err := db.GetLastProcessedBlock(ctx, string(syncTask))
	if err != nil {
		log.Errorf("error retrieving last processed block for %s task, starting from 0: %v", syncTask, err)
	}

	if start > 0 {
		start = start - 1 // since a block may have been partially processed
	}

	return start, err
}

func setStartBlock(parentCtx context.Context, db dbTypes.DB, block uint64, syncTask SyncTask) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.StoreLastProcessedBlock(ctx, block, string(syncTask))
}

func storeUnresolvedBatchKeys(parentCtx context.Context, db dbTypes.DB, keys []types.BatchKey) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.StoreUnresolvedBatchKeys(ctx, keys)
}

func getUnresolvedBatchKeys(parentCtx context.Context, db dbTypes.DB) ([]types.BatchKey, error) {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.GetUnresolvedBatchKeys(ctx, maxUnprocessedBatch)
}

func deleteUnresolvedBatchKeys(parentCtx context.Context, db dbTypes.DB, keys []types.BatchKey) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.DeleteUnresolvedBatchKeys(ctx, keys)
}

func listOffchainData(parentCtx context.Context, db dbTypes.DB, keys []common.Hash) ([]types.OffChainData, error) {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.ListOffChainData(ctx, keys)
}

func storeOffchainData(parentCtx context.Context, db dbTypes.DB, data []types.OffChainData) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.StoreOffChainData(ctx, data)
}

func detectOffchainDataGaps(parentCtx context.Context, db dbTypes.DB) (map[uint64]uint64, error) {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.DetectOffchainDataGaps(ctx)
}
