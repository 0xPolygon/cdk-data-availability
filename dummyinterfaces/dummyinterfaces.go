package dummyinterfaces

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk-validium-node/jsonrpc"
	"github.com/0xPolygon/cdk-validium-node/pool"
	"github.com/0xPolygon/cdk-validium-node/state"
	"github.com/0xPolygon/cdk-validium-node/state/runtime"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v4"
)

const notImplemented = "not implemented"

// POOL INTERFACE

// DummyPool foo
type DummyPool struct{}

// GetGasPrices implements types.PoolInterface.
func (*DummyPool) GetGasPrices(ctx context.Context) (pool.GasPrices, error) {
	panic("unimplemented")
}

// CheckPolicy foo bar
func (d *DummyPool) CheckPolicy(ctx context.Context, policy pool.PolicyName, address common.Address) (bool, error) {
	return false, errors.New(notImplemented)
}

// AddTx foo
func (d *DummyPool) AddTx(ctx context.Context, tx types.Transaction, ip string) error {
	return errors.New(notImplemented)
}

// GetGasPrice foo
func (d *DummyPool) GetGasPrice(ctx context.Context) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetNonce foo
func (d *DummyPool) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetPendingTxHashesSince foo
func (d *DummyPool) GetPendingTxHashesSince(ctx context.Context, since time.Time) ([]common.Hash, error) {
	return nil, errors.New(notImplemented)
}

// GetPendingTxs foo
func (d *DummyPool) GetPendingTxs(ctx context.Context, limit uint64) ([]pool.Transaction, error) {
	return nil, errors.New(notImplemented)
}

// CountPendingTransactions foo
func (d *DummyPool) CountPendingTransactions(ctx context.Context) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetTxByHash foo
func (d *DummyPool) GetTxByHash(ctx context.Context, hash common.Hash) (*pool.Transaction, error) {
	return nil, errors.New(notImplemented)
}

// STORAGE INTERFACE

// DummyStorage foo
type DummyStorage struct{}

// GetAllBlockFiltersWithWSConn foo
func (d *DummyStorage) GetAllBlockFiltersWithWSConn() ([]*jsonrpc.Filter, error) {
	return nil, errors.New(notImplemented)
}

// GetAllLogFiltersWithWSConn foo
func (d *DummyStorage) GetAllLogFiltersWithWSConn() ([]*jsonrpc.Filter, error) {
	return nil, errors.New(notImplemented)
}

// GetFilter foo
func (d *DummyStorage) GetFilter(filterID string) (*jsonrpc.Filter, error) {
	return nil, errors.New(notImplemented)
}

// NewBlockFilter foo
func (d *DummyStorage) NewBlockFilter(wsConn *websocket.Conn) (string, error) {
	return "", errors.New(notImplemented)
}

// NewLogFilter foo
func (d *DummyStorage) NewLogFilter(wsConn *websocket.Conn, filter jsonrpc.LogFilter) (string, error) {
	return "", errors.New(notImplemented)
}

// NewPendingTransactionFilter foo
func (d *DummyStorage) NewPendingTransactionFilter(wsConn *websocket.Conn) (string, error) {
	return "", errors.New(notImplemented)
}

// UninstallFilter foo
func (d *DummyStorage) UninstallFilter(filterID string) error {
	return errors.New(notImplemented)
}

// UninstallFilterByWSConn foo
func (d *DummyStorage) UninstallFilterByWSConn(wsConn *websocket.Conn) error {
	return errors.New(notImplemented)
}

// UpdateFilterLastPoll foo
func (d *DummyStorage) UpdateFilterLastPoll(filterID string) error {
	return errors.New(notImplemented)
}

// STATE INTERFACE METHODS

// DummyState foo
type DummyState struct{}

// PrepareWebSocket foo
func (d *DummyState) PrepareWebSocket() {}

// BeginStateTransaction foo
func (d *DummyState) BeginStateTransaction(ctx context.Context) (pgx.Tx, error) {
	return nil, errors.New(notImplemented)
}

// DebugTransaction foo
func (d *DummyState) DebugTransaction(ctx context.Context, transactionHash common.Hash, traceConfig state.TraceConfig, dbTx pgx.Tx) (*runtime.ExecutionResult, error) {
	return nil, errors.New(notImplemented)
}

// EstimateGas foo
func (d *DummyState) EstimateGas(transaction *types.Transaction, senderAddress common.Address, l2BlockNumber *uint64, dbTx pgx.Tx) (uint64, []byte, error) {
	return 0, nil, errors.New(notImplemented)
}

// GetBalance foo
func (d *DummyState) GetBalance(ctx context.Context, address common.Address, root common.Hash) (*big.Int, error) {
	return nil, errors.New(notImplemented)
}

// GetCode foo
func (d *DummyState) GetCode(ctx context.Context, address common.Address, root common.Hash) ([]byte, error) {
	return nil, errors.New(notImplemented)
}

// GetL2BlockByHash foo
func (d *DummyState) GetL2BlockByHash(ctx context.Context, hash common.Hash, dbTx pgx.Tx) (*types.Block, error) {
	return nil, errors.New(notImplemented)
}

// GetL2BlockByNumber foo
func (d *DummyState) GetL2BlockByNumber(ctx context.Context, blockNumber uint64, dbTx pgx.Tx) (*types.Block, error) {
	return nil, errors.New(notImplemented)
}

// BatchNumberByL2BlockNumber foo
func (d *DummyState) BatchNumberByL2BlockNumber(ctx context.Context, blockNumber uint64, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetL2BlockHashesSince foo
func (d *DummyState) GetL2BlockHashesSince(ctx context.Context, since time.Time, dbTx pgx.Tx) ([]common.Hash, error) {
	return nil, errors.New(notImplemented)
}

// GetL2BlockHeaderByNumber foo
func (d *DummyState) GetL2BlockHeaderByNumber(ctx context.Context, blockNumber uint64, dbTx pgx.Tx) (*types.Header, error) {
	return nil, errors.New(notImplemented)
}

// GetL2BlockTransactionCountByHash foo
func (d *DummyState) GetL2BlockTransactionCountByHash(ctx context.Context, hash common.Hash, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetL2BlockTransactionCountByNumber foo
func (d *DummyState) GetL2BlockTransactionCountByNumber(ctx context.Context, blockNumber uint64, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetLastVirtualizedL2BlockNumber foo
func (d *DummyState) GetLastVirtualizedL2BlockNumber(ctx context.Context, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetLastConsolidatedL2BlockNumber foo
func (d *DummyState) GetLastConsolidatedL2BlockNumber(ctx context.Context, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetLastL2Block foo
func (d *DummyState) GetLastL2Block(ctx context.Context, dbTx pgx.Tx) (*types.Block, error) {
	return nil, errors.New(notImplemented)
}

// GetLastL2BlockNumber foo
func (d *DummyState) GetLastL2BlockNumber(ctx context.Context, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetLogs foo
func (d *DummyState) GetLogs(ctx context.Context, fromBlock uint64, toBlock uint64, addresses []common.Address, topics [][]common.Hash, blockHash *common.Hash, since *time.Time, dbTx pgx.Tx) ([]*types.Log, error) {
	return nil, errors.New(notImplemented)
}

// GetNonce foo
func (d *DummyState) GetNonce(ctx context.Context, address common.Address, root common.Hash) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetStorageAt foo
func (d *DummyState) GetStorageAt(ctx context.Context, address common.Address, position *big.Int, root common.Hash) (*big.Int, error) {
	return nil, errors.New(notImplemented)
}

// GetSyncingInfo foo
func (d *DummyState) GetSyncingInfo(ctx context.Context, dbTx pgx.Tx) (state.SyncingInfo, error) {
	return state.SyncingInfo{}, errors.New(notImplemented)
}

// GetTransactionByHash foo
func (d *DummyState) GetTransactionByHash(ctx context.Context, transactionHash common.Hash, dbTx pgx.Tx) (*types.Transaction, error) {
	return nil, errors.New(notImplemented)
}

// GetTransactionByL2BlockHashAndIndex foo
func (d *DummyState) GetTransactionByL2BlockHashAndIndex(ctx context.Context, blockHash common.Hash, index uint64, dbTx pgx.Tx) (*types.Transaction, error) {
	return nil, errors.New(notImplemented)
}

// GetTransactionByL2BlockNumberAndIndex foo
func (d *DummyState) GetTransactionByL2BlockNumberAndIndex(ctx context.Context, blockNumber uint64, index uint64, dbTx pgx.Tx) (*types.Transaction, error) {
	return nil, errors.New(notImplemented)
}

// GetTransactionReceipt foo
func (d *DummyState) GetTransactionReceipt(ctx context.Context, transactionHash common.Hash, dbTx pgx.Tx) (*types.Receipt, error) {
	return nil, errors.New(notImplemented)
}

// IsL2BlockConsolidated foo
func (d *DummyState) IsL2BlockConsolidated(ctx context.Context, blockNumber uint64, dbTx pgx.Tx) (bool, error) {
	return false, errors.New(notImplemented)
}

// IsL2BlockVirtualized foo
func (d *DummyState) IsL2BlockVirtualized(ctx context.Context, blockNumber uint64, dbTx pgx.Tx) (bool, error) {
	return false, errors.New(notImplemented)
}

// ProcessUnsignedTransaction foo
func (d *DummyState) ProcessUnsignedTransaction(ctx context.Context, tx *types.Transaction, _ common.Address, _ *uint64, _ bool, dbTx pgx.Tx) (*runtime.ExecutionResult, error) {
	return nil, errors.New(notImplemented)
}

// RegisterNewL2BlockEventHandler foo
func (d *DummyState) RegisterNewL2BlockEventHandler(h state.NewL2BlockEventHandler) {}

// GetLastVirtualBatchNum foo
func (d *DummyState) GetLastVirtualBatchNum(ctx context.Context, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetLastVerifiedBatch foo
func (d *DummyState) GetLastVerifiedBatch(ctx context.Context, dbTx pgx.Tx) (*state.VerifiedBatch, error) {
	return nil, errors.New(notImplemented)
}

// GetLastBatchNumber foo
func (d *DummyState) GetLastBatchNumber(ctx context.Context, dbTx pgx.Tx) (uint64, error) {
	return 0, errors.New(notImplemented)
}

// GetBatchByNumber foo
func (d *DummyState) GetBatchByNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.Batch, error) {
	return nil, errors.New(notImplemented)
}

// GetTransactionsByBatchNumber foo
func (d *DummyState) GetTransactionsByBatchNumber(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (txs []types.Transaction, effectivePercentages []uint8, err error) {
	return
}

// GetVirtualBatch foo
func (d *DummyState) GetVirtualBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.VirtualBatch, error) {
	return nil, errors.New(notImplemented)
}

// GetVerifiedBatch foo
func (d *DummyState) GetVerifiedBatch(ctx context.Context, batchNumber uint64, dbTx pgx.Tx) (*state.VerifiedBatch, error) {
	return nil, errors.New(notImplemented)
}

// GetExitRootByGlobalExitRoot foo
func (d *DummyState) GetExitRootByGlobalExitRoot(ctx context.Context, ger common.Hash, dbTx pgx.Tx) (*state.GlobalExitRoot, error) {
	return nil, errors.New(notImplemented)
}
