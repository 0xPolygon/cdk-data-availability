package datacom

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/0xPolygon/cdk-data-availability/types"
)

// DBInterface is the interface needed by the datacom service
type DBInterface interface {
	BeginStateTransaction(ctx context.Context) (*sqlx.Tx, error)
	StoreOffChainData(ctx context.Context, od []types.OffChainData, dbTx *sqlx.Tx) error
}
