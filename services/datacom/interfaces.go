package datacom

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/offchaindata"
	"github.com/jackc/pgx/v4"
)

// DBInterface is the interface needed by the datacom service
type DBInterface interface {
	BeginStateTransaction(ctx context.Context) (pgx.Tx, error)
	StoreOffChainData(ctx context.Context, od []offchaindata.OffChainData, dbTx pgx.Tx) error
}
