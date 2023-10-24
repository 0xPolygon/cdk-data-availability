package synchronizer

import (
	"context"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-data-availability/config"
	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	initBlockTimeout = 15 * time.Second
	minCodeLen       = 2
)

// InitStartBlock initializes the L1 sync task by finding the inception block for the CDKValidium contract
func InitStartBlock(db *db.DB, l1 config.L1Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), initBlockTimeout)
	defer cancel()

	current, err := getStartBlock(db)
	if err != nil {
		return err
	}
	if current > 0 {
		// no need to resolve start block, it's already been set
		return nil
	}
	log.Info("starting search for start block of contract ", l1.CDKValidiumAddress)
	startBlock, err := findContractDeploymentBlock(ctx, l1.RpcURL, common.HexToAddress(l1.CDKValidiumAddress))
	if err != nil {
		return err
	}
	err = setStartBlock(db, startBlock.Uint64())
	if err != nil {
		return err
	}
	return nil
}

func findContractDeploymentBlock(ctx context.Context, url string, contract common.Address) (*big.Int, error) {
	eth, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, err
	}
	latestBlock, err := eth.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	firstBlock := findCode(ctx, eth, contract, 0, latestBlock.Number().Int64())
	return big.NewInt(firstBlock), nil
}

// findCode is an O(log(n)) search for the inception block of a contract at the given address
func findCode(ctx context.Context, eth *ethclient.Client, address common.Address, startBlock, endBlock int64) int64 {
	if startBlock == endBlock {
		return startBlock
	}
	midBlock := (startBlock + endBlock) / 2 //nolint:gomnd
	if codeLen := codeLen(ctx, eth, address, midBlock); codeLen > minCodeLen {
		return findCode(ctx, eth, address, startBlock, midBlock)
	} else {
		return findCode(ctx, eth, address, midBlock+1, endBlock)
	}
}

func codeLen(ctx context.Context, eth *ethclient.Client, address common.Address, blockNumber int64) int64 {
	data, err := eth.CodeAt(ctx, address, big.NewInt(blockNumber))
	if err != nil {
		return 0
	}
	return int64(len(data))
}
