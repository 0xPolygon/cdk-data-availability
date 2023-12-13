package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// EthClientFactory defines functions for a EthClient factory
type EthClientFactory interface {
	CreateEthClient(ctx context.Context, url string) (EthClient, error)
}

// EthClient defines functions that an ethereum rpc client should implement
type EthClient interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}
