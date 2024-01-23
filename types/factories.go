package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// IEthClient defines functions that an ethereum rpc client should implement
type IEthClient interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}

// IEthClientFactory defines functions for a EthClient factory
type IEthClientFactory interface {
	CreateEthClient(ctx context.Context, url string) (IEthClient, error)
}

var _ IEthClientFactory = (*EthClientFactoryImpl)(nil)

// EthClientFactoryImpl is the implementation of EthClientFactory interface
type EthClientFactoryImpl struct{}

// CreateEthClient creates a new eth client
func (e *EthClientFactoryImpl) CreateEthClient(ctx context.Context, url string) (IEthClient, error) {
	return ethclient.DialContext(ctx, url)
}
