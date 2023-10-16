package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/common"
)

// GetOffChainData returns data based on it's hash
func (c *Client) GetOffChainData(ctx context.Context, hash common.Hash) ([]byte, error) {
	response, err := rpc.JSONRPCCall(c.url, "sync_getOffChainData", hash)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result rpc.ArgBytes
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
