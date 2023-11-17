package sequencer

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
)

// GetData returns batch data from the trusted sequencer
func GetData(url string, batchNum uint64) (*types.Batch, error) {
	response, err := rpc.JSONRPCCall(url, "zkevm_getBatchByNumber", batchNum, true)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%d - %s", response.Error.Code, response.Error.Message)
	}
	var result types.Batch
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
