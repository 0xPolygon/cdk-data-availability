package sync

import (
	"context"

	"github.com/0xPolygon/supernets2-data-availability/jsonrpc/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

// APISYNC  is the namespace of the sync service
const APISYNC = "sync"

// DBInterface is the interface needed by the sync service
type DBInterface interface {
	BeginStateTransaction(ctx context.Context) (pgx.Tx, error)
	GetOffChainData(ctx context.Context, key common.Hash, dbTx pgx.Tx) (types.ArgBytes, error)
}
