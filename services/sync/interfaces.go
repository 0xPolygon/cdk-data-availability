package sync

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/common"
)

// DBInterface is the interface needed by the sync service
type DBInterface interface {
	BeginStateTransaction(ctx context.Context) (*sqlx.Tx, error)
	GetOffChainData(ctx context.Context, key common.Hash, dbTx *sqlx.Tx) (rpc.ArgBytes, error)
}
