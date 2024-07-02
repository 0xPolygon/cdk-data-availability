package sync

import (
	"context"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// APISYNC  is the namespace of the sync service
	APISYNC = "sync"

	// maxListHashes is the maximum number of hashes that can be requested in a ListOffChainData call
	maxListHashes = 100
)

// Endpoints contains implementations for the "zkevm" RPC endpoints
type Endpoints struct {
	db db.DB
}

// NewEndpoints returns Endpoints
func NewEndpoints(db db.DB) *Endpoints {
	return &Endpoints{
		db: db,
	}
}

// GetOffChainData returns the image of the given hash
func (z *Endpoints) GetOffChainData(hash types.ArgHash) (interface{}, rpc.Error) {
	data, err := z.db.GetOffChainData(context.Background(), hash.Hash())
	if err != nil {
		log.Errorf("failed to get the offchain requested data from the DB: %v", err)
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, "failed to get the requested data")
	}

	return types.ArgBytes(data.Value), nil
}

// ListOffChainData returns the list of images of the given hashes
func (z *Endpoints) ListOffChainData(hashes []types.ArgHash) (interface{}, rpc.Error) {
	if len(hashes) > maxListHashes {
		log.Errorf("too many hashes requested in ListOffChainData: %d", len(hashes))
		return "0x0", rpc.NewRPCError(rpc.InvalidRequestErrorCode, "too many hashes requested")
	}

	keys := make([]common.Hash, len(hashes))
	for i, hash := range hashes {
		keys[i] = hash.Hash()
	}

	list, err := z.db.ListOffChainData(context.Background(), keys)
	if err != nil {
		log.Errorf("failed to list the requested data from the DB: %v", err)
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, "failed to list the requested data")
	}

	listMap := make(map[common.Hash]types.ArgBytes)
	for _, data := range list {
		listMap[data.Key] = data.Value
	}

	return listMap, nil
}
