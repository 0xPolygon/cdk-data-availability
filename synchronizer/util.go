package synchronizer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	elderberryValidium "github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonvalidiumetrog"
	etrogValidium "github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonvalidiumetrog"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	// methodIDSequenceBatchesValidiumEtrog is sequenceBatchesValidium method id in Etrog fork (0x2d72c248)
	methodIDSequenceBatchesValidiumEtrog = crypto.Keccak256(
		[]byte("sequenceBatchesValidium((bytes32,bytes32,uint64,bytes32)[],address,bytes)"),
	)[:methodIDLen]

	// methodIDSequenceBatchesValidiumElderberry is sequenceBatchesValidium method id in Elderberry fork (0xdb5b0ed7)
	methodIDSequenceBatchesValidiumElderberry = crypto.Keccak256(
		[]byte("sequenceBatchesValidium((bytes32,bytes32,uint64,bytes32)[],uint64,uint64,address,bytes)"),
	)[:methodIDLen]
)

const (
	// methodIDLen represents method id size in bytes
	methodIDLen = 4
)

// UnpackTxData unpacks the keys in a SequenceBatches event
func UnpackTxData(txData []byte) ([]common.Hash, error) {
	methodID := txData[:methodIDLen]

	var (
		a   abi.ABI
		err error
	)

	if bytes.Equal(methodID, methodIDSequenceBatchesValidiumEtrog) {
		a, err = abi.JSON(strings.NewReader(etrogValidium.PolygonvalidiumetrogMetaData.ABI))
		if err != nil {
			return nil, err
		}
	} else if bytes.Equal(methodID, methodIDSequenceBatchesValidiumElderberry) {
		a, err = abi.JSON(strings.NewReader(elderberryValidium.PolygonvalidiumetrogMetaData.ABI))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unrecognized method id: %s", hex.EncodeToString(methodID))
	}

	method, err := a.MethodById(methodID)
	if err != nil {
		return nil, err
	}

	data, err := method.Inputs.Unpack(txData[methodIDLen:])
	if err != nil {
		return nil, err
	}

	bytes, err := json.Marshal(data[0])
	if err != nil {
		return nil, err
	}

	var batches []etrogValidium.PolygonValidiumEtrogValidiumBatchData
	if err = json.Unmarshal(bytes, &batches); err != nil {
		return nil, err
	}

	keys := make([]common.Hash, len(batches))
	for i, batch := range batches {
		keys[i] = batch.TransactionsHash
	}
	return keys, nil
}
