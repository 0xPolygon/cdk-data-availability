package sync

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// APISYNC  is the namespace of the sync service
const APISYNC = "sync"

// Endpoints contains implementations for the "zkevm" RPC endpoints
type Endpoints struct {
	db    db.DB
	txMan rpc.DBTxManager
}

// NewEndpoints returns Endpoints
func NewEndpoints(db db.DB) *Endpoints {
	return &Endpoints{
		db: db,
	}
}

// GetOffChainData returns the image of the given hash
func (z *Endpoints) GetOffChainData(hash types.ArgHash) (interface{}, rpc.Error) {
	return z.txMan.NewDbTxScope(z.db, func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
		data, err := z.db.GetOffChainData(ctx, hash.Hash(), dbTx)
		if err != nil {
			log.Errorf("failed to get the offchain requested data from the DB: %v", err)
			return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, "failed to get the requested data")
		}

		return data, nil
	})
}

// ListOffChainData returns the list of images of the given hashes
func (z *Endpoints) ListOffChainData(hashes []types.ArgHash) (interface{}, rpc.Error) {
	keys := make([]common.Hash, len(hashes))
	for i, hash := range hashes {
		keys[i] = hash.Hash()
	}

	return z.txMan.NewDbTxScope(z.db, func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
		list, err := z.db.ListOffChainData(ctx, keys, dbTx)
		if err != nil {
			log.Errorf("failed to list the requested data from the DB: %v", err)
			return nil, rpc.NewRPCError(rpc.DefaultErrorCode, "failed to list the requested data")
		}

		return list, nil
	})
}
