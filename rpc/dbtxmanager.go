package rpc

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/db"
)

// DBTxManager allows to do scopped DB txs
type DBTxManager struct{}

// DBTxScopedFn function to do scopped DB txs
type DBTxScopedFn func(ctx context.Context, dbTx db.Tx) (interface{}, Error)

// NewDbTxScope function to initiate DB scopped txs
func (f *DBTxManager) NewDbTxScope(db db.DB, scopedFn DBTxScopedFn) (interface{}, Error) {
	ctx := context.Background()

	dbTx, err := db.BeginStateTransaction(ctx)
	if err != nil {
		return RPCErrorResponse(DefaultErrorCode, "failed to connect to the state", err)
	}

	v, rpcErr := scopedFn(ctx, dbTx)
	if rpcErr != nil {
		if txErr := dbTx.Rollback(); txErr != nil {
			return RPCErrorResponse(DefaultErrorCode, "failed to rollback db transaction", txErr)
		}
		return v, rpcErr
	}

	if txErr := dbTx.Commit(); txErr != nil {
		return RPCErrorResponse(DefaultErrorCode, "failed to commit db transaction", txErr)
	}
	return v, rpcErr
}
