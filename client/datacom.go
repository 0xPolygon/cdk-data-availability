package client

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
)

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

	var result rpc.ArgBytes
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
