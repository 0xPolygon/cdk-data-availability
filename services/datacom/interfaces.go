package datacom

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/jmoiron/sqlx"
)

// DBInterface is the interface needed by the datacom service
type DBInterface interface {
	BeginStateTransaction(ctx context.Context) (*sqlx.Tx, error)
	StoreOffChainData(ctx context.Context, od []types.OffChainData, dbTx *sqlx.Tx) error
}
