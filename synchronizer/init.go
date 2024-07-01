package synchronizer

import (
	"context"
	"math/big"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/common"
)

const (
	initBlockTimeout    = 15 * time.Second
	minCodeLen          = 2
	maxUnprocessedBatch = 100
)

// InitStartBlock initializes the L1 sync task by finding the inception block for the CDKValidium contract
func InitStartBlock(parentCtx context.Context, db db.DB, em etherman.Etherman, genesisBlock uint64, validiumAddr common.Address) error {
	ctx, cancel := context.WithTimeout(parentCtx, initBlockTimeout)
	defer cancel()

	// Get start block either from genesis block or from contract deployment block number
	startBlock, err := getInitialStartBlock(ctx, em, genesisBlock, validiumAddr)
	if err != nil {
		return err
	}

	initializeBlock := func(syncTask SyncTask) error {
		current, err := getStartBlock(ctx, db, syncTask)
		if err != nil {
			return err
		}

		if current > 0 {
			// no need to resolve start block, it's already been set
			return nil
		}

		return setStartBlock(ctx, db, startBlock.Uint64(), syncTask)
	}

	var errGroup errgroup.Group

	// Init start block for L1SyncTask
	errGroup.Go(func() error {
		log.Info("initializing start block for L1 sync task")

		return initializeBlock(L1SyncTask)
	})

	// Init start block for L1BatchNumTask
	errGroup.Go(func() error {
		log.Info("initializing start block for L1_BATCH_NUM sync task")

		return initializeBlock(L1BatchNumTask)
	})

	return errGroup.Wait()
}

func getInitialStartBlock(
	ctx context.Context,
	em etherman.Etherman,
	genesisBlock uint64,
	validiumAddr common.Address,
) (*big.Int, error) {
	startBlock := new(big.Int)
	if genesisBlock != 0 {
		startBlock.SetUint64(genesisBlock)
	} else {
		var err error
		if startBlock, err = findContractDeploymentBlock(ctx, em, validiumAddr); err != nil {
			return nil, err
		}
	}

	return startBlock, nil
}

func findContractDeploymentBlock(ctx context.Context, em etherman.Etherman, contract common.Address) (*big.Int, error) {
	latestHeader, err := em.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	firstBlock := findCode(ctx, em, contract, 0, latestHeader.Number.Int64())
	return big.NewInt(firstBlock), nil
}

// findCode is an O(log(n)) search for the inception block of a contract at the given address
func findCode(ctx context.Context, em etherman.Etherman, address common.Address, startBlock, endBlock int64) int64 {
	if startBlock == endBlock {
		return startBlock
	}
	midBlock := (startBlock + endBlock) / 2 //nolint:gomnd
	if codeLen := codeLen(ctx, em, address, midBlock); codeLen > minCodeLen {
		return findCode(ctx, em, address, startBlock, midBlock)
	} else {
		return findCode(ctx, em, address, midBlock+1, endBlock)
	}
}

func codeLen(ctx context.Context, em etherman.Etherman, address common.Address, blockNumber int64) int64 {
	data, err := em.CodeAt(ctx, address, big.NewInt(blockNumber))
	if err != nil {
		return 0
	}
	return int64(len(data))
}
