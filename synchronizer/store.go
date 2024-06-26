package synchronizer

import (
	"context"
	"time"

	dbTypes "github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

const dbTimeout = 2 * time.Second

// L1SyncTask is the name of the L1 sync task
const L1SyncTask = "L1"

func getStartBlock(parentCtx context.Context, db dbTypes.DB) (uint64, error) {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	start, err := db.GetLastProcessedBlock(ctx, L1SyncTask)
	if err != nil {
		log.Errorf("error retrieving last processed block, starting from 0: %v", err)
	}

	if start > 0 {
		start = start - 1 // since a block may have been partially processed
	}

	return start, err
}

func setStartBlock(parentCtx context.Context, db dbTypes.DB, block uint64) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.StoreLastProcessedBlock(ctx, L1SyncTask, block)
}

func exists(parentCtx context.Context, db dbTypes.DB, key common.Hash) bool {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.Exists(ctx, key)
}

func storeUnresolvedBatchKeys(parentCtx context.Context, db dbTypes.DB, keys []types.BatchKey) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.StoreUnresolvedBatchKeys(ctx, keys)
}

func getUnresolvedBatchKeys(parentCtx context.Context, db dbTypes.DB) ([]types.BatchKey, error) {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.GetUnresolvedBatchKeys(ctx)
}

func deleteUnresolvedBatchKeys(parentCtx context.Context, db dbTypes.DB, keys []types.BatchKey) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.DeleteUnresolvedBatchKeys(ctx, keys)
}

func storeOffchainData(parentCtx context.Context, db dbTypes.DB, data []types.OffChainData) error {
	ctx, cancel := context.WithTimeout(parentCtx, dbTimeout)
	defer cancel()

	return db.StoreOffChainData(ctx, data)
}
