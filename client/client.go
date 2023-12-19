package client

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// IClientFactory interface for the client factory
//
//go:generate mockery --name IClientFactory --output ../mocks --case=underscore --filename client_factory.generated.go
type IClientFactory interface {
	New(url string) IClient
}

// IClient is the interface that defines the implementation of all the endpoints
//
//go:generate mockery --name IClient --output ../mocks --case=underscore --filename client.generated.go
type IClient interface {
	GetOffChainData(ctx context.Context, hash common.Hash) ([]byte, error)
	SignSequence(signedSequence types.SignedSequence) ([]byte, error)
}

// ClientFactory is the implementation of the data committee client factory
type ClientFactory struct{}

// New returns an implementation of the data committee node client
func (f *ClientFactory) New(url string) IClient {
	return New(url)
}

// Client wraps all the available endpoints of the data abailability committee node server
type Client struct {
	url string
}

// New returns a client ready to be used
func New(url string) *Client {
	return &Client{
		url: url,
	}
}
