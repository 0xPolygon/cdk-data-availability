package sequencer

import (
	"encoding/json"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
)

// GetData returns batch data from the trusted sequencer
func GetData(url string, batchNum uint64) (*SequenceBatch, error) {
	response, err := rpc.JSONRPCCall(url, "zkevm_getBatchByNumber", batchNum, true)
	if err != nil {
		return nil, err
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%d - %s", response.Error.Code, response.Error.Message)
	}
	var result SequenceBatch
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SequenceBatch is the batch by number structure returned by trusted sequencer
type SequenceBatch struct {
	Number              rpc.ArgUint64  `json:"number"`
	ForcedBatchNumber   *rpc.ArgUint64 `json:"forcedBatchNumber,omitempty"`
	Coinbase            common.Address `json:"coinbase"`
	StateRoot           common.Hash    `json:"stateRoot"`
	GlobalExitRoot      common.Hash    `json:"globalExitRoot"`
	MainnetExitRoot     common.Hash    `json:"mainnetExitRoot"`
	RollupExitRoot      common.Hash    `json:"rollupExitRoot"`
	LocalExitRoot       common.Hash    `json:"localExitRoot"`
	AccInputHash        common.Hash    `json:"accInputHash"`
	Timestamp           rpc.ArgUint64  `json:"timestamp"`
	SendSequencesTxHash *common.Hash   `json:"sendSequencesTxHash"`
	VerifyBatchTxHash   *common.Hash   `json:"verifyBatchTxHash"`
	Closed              bool           `json:"closed"`
	//Blocks              []BlockOrHash       `json:"blocks"`
	Transactions []TransactionOrHash `json:"transactions"`
	BatchL2Data  rpc.ArgBytes        `json:"batchL2Data"`
}

// TransactionOrHash for union type of transaction and types.Hash
type TransactionOrHash struct {
	Hash *common.Hash
	Tx   *Transaction
}

// Transaction structure
type Transaction struct {
	Nonce       rpc.ArgUint64   `json:"nonce"`
	GasPrice    rpc.ArgBig      `json:"gasPrice"`
	Gas         rpc.ArgUint64   `json:"gas"`
	To          *common.Address `json:"to"`
	Value       rpc.ArgBig      `json:"value"`
	Input       rpc.ArgBytes    `json:"input"`
	V           rpc.ArgBig      `json:"v"`
	R           rpc.ArgBig      `json:"r"`
	S           rpc.ArgBig      `json:"s"`
	Hash        common.Hash     `json:"hash"`
	From        common.Address  `json:"from"`
	BlockHash   *common.Hash    `json:"blockHash"`
	BlockNumber *rpc.ArgUint64  `json:"blockNumber"`
	TxIndex     *rpc.ArgUint64  `json:"transactionIndex"`
	ChainID     rpc.ArgBig      `json:"chainId"`
	Type        rpc.ArgUint64   `json:"type"`
	Receipt     *Receipt        `json:"receipt,omitempty"`
}

// Receipt structure
type Receipt struct {
	Root              common.Hash     `json:"root"`
	CumulativeGasUsed rpc.ArgUint64   `json:"cumulativeGasUsed"`
	LogsBloom         etypes.Bloom    `json:"logsBloom"`
	Logs              []*etypes.Log   `json:"logs"`
	Status            rpc.ArgUint64   `json:"status"`
	TxHash            common.Hash     `json:"transactionHash"`
	TxIndex           rpc.ArgUint64   `json:"transactionIndex"`
	BlockHash         common.Hash     `json:"blockHash"`
	BlockNumber       rpc.ArgUint64   `json:"blockNumber"`
	GasUsed           rpc.ArgUint64   `json:"gasUsed"`
	FromAddr          common.Address  `json:"from"`
	ToAddr            *common.Address `json:"to"`
	ContractAddress   *common.Address `json:"contractAddress"`
	Type              rpc.ArgUint64   `json:"type"`
	EffectiveGasPrice *rpc.ArgBig     `json:"effectiveGasPrice,omitempty"`
}
