package synchronizer

import (
	"context"
	"time"

	"github.com/0xPolygon/cdk-data-availability/client"
	"github.com/0xPolygon/cdk-data-availability/etherman"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/offchaindata"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/ethereum/go-ethereum/common"
)

const rpcTimeout = 3 * time.Second

func resolveWithMember(key common.Hash, member etherman.DataCommitteeMember) (offchaindata.OffChainData, error) {
	cm := client.New(member.URL)
	ctx, cancel := context.WithTimeout(context.Background(), rpcTimeout)
	defer cancel()

	log.Debugf("trying member %v at %v for key %v", member.Addr.Hex(), member.URL, key.Hex())

	bytes, err := cm.GetOffChainData(ctx, key)
	if len(bytes) == 0 {
		err = rpc.NewRPCError(rpc.NotFoundErrorCode, "data not found")
	}
	var data offchaindata.OffChainData
	if len(bytes) > 0 {
		data = offchaindata.OffChainData{
			Key:   key,
			Value: bytes,
		}
	}
	return data, err
}
