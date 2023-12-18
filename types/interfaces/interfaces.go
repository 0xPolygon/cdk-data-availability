package interfaces

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// IEthClientFactory defines functions for a EthClient factory
type IEthClientFactory interface {
	CreateEthClient(ctx context.Context, url string) (IEthClient, error)
}

// IEthClient defines functions that an ethereum rpc client should implement
type IEthClient interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}
