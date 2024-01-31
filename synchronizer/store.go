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

const l1SyncTask = "L1"

func getStartBlock(db dbTypes.DB) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	start, err := db.GetLastProcessedBlock(ctx, l1SyncTask)
	if err != nil {
		log.Errorf("error retrieving last processed block, starting from 0: %v", err)
	}
	if start > 0 {
		start = start - 1 // since a block may have been partially processed
	}
	return start, err
}

func setStartBlock(db dbTypes.DB, block uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var (
		dbTx dbTypes.Tx
		err  error
	)

	if dbTx, err = db.BeginStateTransaction(ctx); err != nil {
		return err
	}

	if err = db.StoreLastProcessedBlock(ctx, l1SyncTask, block, dbTx); err != nil {
		return err
	}

	if err = dbTx.Commit(); err != nil {
		return err
	}

	return nil
}

func exists(db dbTypes.DB, key common.Hash) bool {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	return db.Exists(ctx, key)
}

func store(db dbTypes.DB, data []types.OffChainData) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var (
		dbTx dbTypes.Tx
		err  error
	)

	if dbTx, err = db.BeginStateTransaction(ctx); err != nil {
		return err
	}

	if err = db.StoreOffChainData(ctx, data, dbTx); err != nil {
		rollback(err, dbTx)
		return err
	}

	if err = dbTx.Commit(); err != nil {
		return err
	}

	return nil
}

func rollback(err error, dbTx dbTypes.Tx) {
	if txErr := dbTx.Rollback(); txErr != nil {
		log.Errorf("failed to roll back transaction after error %v : %v", err, txErr)
	}
}
