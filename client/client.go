package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
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

// Factory is the implementation of the data committee client factory
type Factory struct{}

// New returns an implementation of the data committee node client
func (f *Factory) New(url string) IClient {
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

// SignSequence sends a request to sign the given sequence by the data committee member
// if successful returns the signature. The signature should be validated after using this method!
func (c *Client) SignSequence(signedSequence types.SignedSequence) ([]byte, error) {
	response, err := rpc.JSONRPCCall(c.url, "datacom_signSequence", signedSequence)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result types.ArgBytes
	if err = json.Unmarshal(response.Result, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetOffChainData returns data based on it's hash
func (c *Client) GetOffChainData(ctx context.Context, hash common.Hash) ([]byte, error) {
	response, err := rpc.JSONRPCCallWithContext(ctx, c.url, "sync_getOffChainData", hash)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result types.ArgBytes
	if err = json.Unmarshal(response.Result, &result); err != nil {
		return nil, err
	}

	return result, nil
}
