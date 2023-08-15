package sync

import (
	"context"

	"github.com/0xPolygon/supernets2-data-availability/jsonrpc"
	"github.com/0xPolygon/supernets2-data-availability/jsonrpc/types"
	"github.com/jackc/pgx/v4"
)

// Endpoints contains implementations for the "sync" RPC endpoints
type Endpoints struct {
	db    DBInterface
	txMan jsonrpc.DBTxManager
}

// NewEndpoints returns ZKEVMEndpoints
func NewEndpoints(db DBInterface) *Endpoints {
	return &Endpoints{
		db: db,
	}
}

// GetOffChainData returns the image of the given hash
func (z *Endpoints) GetOffChainData(hash types.ArgHash) (interface{}, types.Error) {
	return z.txMan.NewDbTxScope(z.db, func(ctx context.Context, dbTx pgx.Tx) (interface{}, types.Error) {
		data, err := z.db.GetOffChainData(ctx, hash.Hash(), dbTx)
		if err != nil {
			return "0x0", types.NewRPCError(types.DefaultErrorCode, "failed to get the requested data")
		}

		return data, nil
	})
}
