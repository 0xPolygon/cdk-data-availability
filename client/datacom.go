package client

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/supernets2-data-availability/sequence"
	"github.com/0xPolygon/supernets2-node/jsonrpc/client"
	"github.com/0xPolygon/supernets2-node/jsonrpc/types"
)

// SignSequence sends a request to sign the given sequence by the data committee member
// if successful returns the signature. The signature should be validated after using this method!
func (c *Client) SignSequence(signedSequence sequence.SignedSequence) ([]byte, error) {
	response, err := client.JSONRPCCall(c.url, "datacom_signSequence", signedSequence)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%v %v", response.Error.Code, response.Error.Message)
	}

	var result types.ArgBytes
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
