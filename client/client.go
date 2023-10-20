package client

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// ClientFactoryInterface interface for the client factory
type ClientFactoryInterface interface {
	New(url string) ClientInterface
}

// ClientInterface is the interface that defines the implementation of all the endpoints
type ClientInterface interface {
	GetOffChainData(ctx context.Context, hash common.Hash) ([]byte, error)
	SignSequence(signedSequence types.SignedSequence) ([]byte, error)
}

// ClientFactory is the implementation of the data committee client factory
type ClientFactory struct{}

// New returns an implementation of the data committee node client
func (f *ClientFactory) New(url string) ClientInterface {
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
