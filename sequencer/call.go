package sequencer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// SeqBatch structure
type SeqBatch struct {
	Number       types.ArgUint64 `json:"number"`
	AccInputHash common.Hash     `json:"accInputHash"`
	BatchL2Data  types.ArgBytes  `json:"batchL2Data"`
}

// GetData returns batch data from the trusted sequencer
func GetData(ctx context.Context, url string, batchNum uint64) (*SeqBatch, error) {
	response, err := rpc.JSONRPCCallWithContext(ctx, url, "zkevm_getBatchByNumber", batchNum, true)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, fmt.Errorf("%d - %s", response.Error.Code, response.Error.Message)
	}

	var result SeqBatch
	if err = json.Unmarshal(response.Result, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
