package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient defines functions that an ethereum rpc client should implement
type EthClient interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}

// EthClientFactory defines functions for a EthClient factory
type EthClientFactory interface {
	CreateEthClient(ctx context.Context, url string) (EthClient, error)
}

// ethClientFactory is the implementation of EthClientFactory interface
type ethClientFactory struct{}

// NewEthClientFactory is the constructor of ethClientFactory
func NewEthClientFactory() EthClientFactory {
	return &ethClientFactory{}
}

// CreateEthClient creates a new eth client
func (e *ethClientFactory) CreateEthClient(ctx context.Context, url string) (EthClient, error) {
	return ethclient.DialContext(ctx, url)
}
