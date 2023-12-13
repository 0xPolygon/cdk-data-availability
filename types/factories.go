package types

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/types/interfaces"
	"github.com/ethereum/go-ethereum/ethclient"
)

var _ interfaces.EthClientFactory = (*EthClientFactoryImpl)(nil)

// EthClientFactoryImpl is the implementation of EthClientFactory interface
type EthClientFactoryImpl struct{}

// CreateEthClient creates a new eth client
func (e *EthClientFactoryImpl) CreateEthClient(ctx context.Context, url string) (interfaces.EthClient, error) {
	return ethclient.DialContext(ctx, url)
}
