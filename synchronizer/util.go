package synchronizer

import (
	"encoding/json"
	"strings"

	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/cdkvalidium"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// UnpackTxData unpacks the keys in a SequenceBatches event
func UnpackTxData(txData []byte) ([]common.Hash, error) {
	a, err := abi.JSON(strings.NewReader(cdkvalidium.CdkvalidiumABI))
	if err != nil {
		return nil, err
	}
	method, err := a.MethodById(txData[:4])
	if err != nil {
		return nil, err
	}
	data, err := method.Inputs.Unpack(txData[4:])
	if err != nil {
		return nil, err
	}
	var batches []cdkvalidium.CDKValidiumBatchData
	bytes, err := json.Marshal(data[0])
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &batches)
	if err != nil {
		return nil, err
	}

	var keys []common.Hash
	for _, batch := range batches {
		keys = append(keys, batch.TransactionsHash)
	}
	return keys, nil
}
