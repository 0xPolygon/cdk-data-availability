package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// Factory interface for the client factory
type Factory interface {
	New(url string) Client
}

// Client is the interface that defines the implementation of all the endpoints
type Client interface {
	GetOffChainData(ctx context.Context, hash common.Hash) ([]byte, error)
	SignSequence(signedSequence types.SignedSequence) ([]byte, error)
}

// factory is the implementation of the data committee client factory
type factory struct{}

// NewFactory is the constructor of factory
func NewFactory() Factory {
	return &factory{}
}

// New returns an implementation of the data committee node client
func (f *factory) New(url string) Client {
	return New(url)
}

// Client wraps all the available endpoints of the data abailability committee node server
type client struct {
	url string
}

// New returns a client ready to be used
func New(url string) Client {
	return &client{
		url: url,
	}
}

// SignSequence sends a request to sign the given sequence by the data committee member
// if successful returns the signature. The signature should be validated after using this method!
func (c *client) SignSequence(signedSequence types.SignedSequence) ([]byte, error) {
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
func (c *client) GetOffChainData(ctx context.Context, hash common.Hash) ([]byte, error) {
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
