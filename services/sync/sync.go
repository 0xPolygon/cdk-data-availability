package sync

import (
	"context"

	"github.com/0xPolygonHermez/zkevm-node/jsonrpc"
	"github.com/0xPolygonHermez/zkevm-node/jsonrpc/types"
	"github.com/jackc/pgx/v4"
)

// APISYNC  is the namespace of the sync service
const APISYNC = "sync"

// SyncEndpoints contains implementations for the "zkevm" RPC endpoints
type SyncEndpoints struct {
	db    DBInterface
	txMan jsonrpc.DBTxManager
}

// NewSyncEndpoints returns ZKEVMEndpoints
func NewSyncEndpoints(db DBInterface) *SyncEndpoints {
	return &SyncEndpoints{
		db: db,
	}
}

// GetOffChainData returns the image of the given hash
func (z *SyncEndpoints) GetOffChainData(hash types.ArgHash) (interface{}, types.Error) {
	return z.txMan.NewDbTxScope(z.db, func(ctx context.Context, dbTx pgx.Tx) (interface{}, types.Error) {
		data, err := z.db.GetOffChainData(ctx, hash.Hash(), dbTx)
		if err != nil {
			return "0x0", types.NewRPCError(types.DefaultErrorCode, "failed to get the requested data")
		}

		return data, nil
	})
}
