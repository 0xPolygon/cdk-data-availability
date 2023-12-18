package sync

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
)

// APISYNC  is the namespace of the sync service
const APISYNC = "sync"

// SyncEndpoints contains implementations for the "zkevm" RPC endpoints
type SyncEndpoints struct {
	db    db.IDB
	txMan rpc.DBTxManager
}

// NewSyncEndpoints returns ZKEVMEndpoints
func NewSyncEndpoints(db db.IDB) *SyncEndpoints {
	return &SyncEndpoints{
		db: db,
	}
}

// GetOffChainData returns the image of the given hash
func (z *SyncEndpoints) GetOffChainData(hash types.ArgHash) (interface{}, rpc.Error) {
	return z.txMan.NewDbTxScope(z.db, func(ctx context.Context, dbTx db.IDBTx) (interface{}, rpc.Error) {
		data, err := z.db.GetOffChainData(ctx, hash.Hash(), dbTx)
		if err != nil {
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, "failed to get the requested data")
		}

		return data, nil
	})
}
