package synchronizer

import (
	"encoding/json"
	"strings"

	"github.com/0xPolygon/cdk-data-availability/etherman/smartcontracts/polygonvalidium"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// UnpackTxData unpacks the keys in a SequenceBatches event
func UnpackTxData(txData []byte) ([]common.Hash, error) {
	a, err := abi.JSON(strings.NewReader(polygonvalidium.PolygonvalidiumMetaData.ABI))
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

	var batches []polygonvalidium.PolygonValidiumEtrogValidiumBatchData
	bytes, err := json.Marshal(data[0])
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &batches); err != nil {
		return nil, err
	}

	keys := make([]common.Hash, len(batches))
	for i, batch := range batches {
		keys[i] = batch.TransactionsHash
	}
	return keys, nil
}
