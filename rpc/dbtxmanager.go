package rpc

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// DBTxManager allows to do scopped DB txs
type DBTxManager struct{}

// DBTxScopedFn function to do scopped DB txs
type DBTxScopedFn func(ctx context.Context, dbTx *sqlx.Tx) (interface{}, Error)

// DBTxer interface to begin DB txs
type DBTxer interface {
	BeginStateTransaction(ctx context.Context) (*sqlx.Tx, error)
}

// NewDbTxScope function to initiate DB scopped txs
func (f *DBTxManager) NewDbTxScope(db DBTxer, scopedFn DBTxScopedFn) (interface{}, Error) {
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
