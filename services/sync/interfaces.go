package sync

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
)

// DBInterface is the interface needed by the sync service
type DBInterface interface {
	BeginStateTransaction(ctx context.Context) (*sqlx.Tx, error)
	GetOffChainData(ctx context.Context, key common.Hash, dbTx sqlx.QueryerContext) (rpc.ArgBytes, error)
}
