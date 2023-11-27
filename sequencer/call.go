package sequencer

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/common"
)

// GetData returns batch data from the trusted sequencer
func GetData(url string, batchNum uint64) (*SeqBatch, error) {
	response, err := rpc.JSONRPCCall(url, "zkevm_getBatchByNumber", batchNum, true)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%d - %s", response.Error.Code, response.Error.Message)
	}
	var result SeqBatch
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SeqBatch structure
type SeqBatch struct {
	Number       rpc.ArgUint64 `json:"number"`
	AccInputHash common.Hash   `json:"accInputHash"`
	BatchL2Data  rpc.ArgBytes  `json:"batchL2Data"`
}
