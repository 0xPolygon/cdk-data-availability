package datacom

import (
	"context"
	"crypto/ecdsa"

	"github.com/0xPolygon/supernets2-data-availability/jsonrpc"
	"github.com/0xPolygon/supernets2-data-availability/jsonrpc/types"
	"github.com/0xPolygon/supernets2-data-availability/sequence"
	"github.com/0xPolygon/supernets2-data-availability/synchronizer"
	"github.com/jackc/pgx/v4"
)

// APIDATACOM is the namespace of the datacom service
const APIDATACOM = "datacom"

// Endpoints contains implementations for the "datacom" RPC endpoints
type Endpoints struct {
	db               DBInterface
	txMan            jsonrpc.DBTxManager
	privateKey       *ecdsa.PrivateKey
	sequencerTracker *synchronizer.SequencerTracker
}

// NewEndpoints returns DataComEndpoints
func NewEndpoints(
	db DBInterface, privateKey *ecdsa.PrivateKey, sequencerTracker *synchronizer.SequencerTracker,
) *Endpoints {
	return &Endpoints{
		db:               db,
		privateKey:       privateKey,
		sequencerTracker: sequencerTracker,
	}
}

// SignSequence generates the accumulated input hash aka accInputHash of the sequence and sign it.
// After storing the data that will be sent hashed to the contract, it returns the signature.
// This endpoint is only accessible to the sequencer
func (d *Endpoints) SignSequence(signedSequence sequence.SignedSequence) (interface{}, types.Error) {
	// Verify that the request comes from the sequencer
	sender, err := signedSequence.Signer()
	if err != nil {
		return "0x0", types.NewRPCError(types.DefaultErrorCode, "failed to verify sender")
	}
	if sender != d.sequencerTracker.GetAddr() {
		return "0x0", types.NewRPCError(types.DefaultErrorCode, "unauthorized")
	}
	// Store off-chain data by hash (hash(L2Data): L2Data)
	_, err = d.txMan.NewDbTxScope(d.db, func(ctx context.Context, dbTx pgx.Tx) (interface{}, types.Error) {
		err := d.db.StoreOffChainData(ctx, signedSequence.Sequence.OffChainData(), dbTx)
		if err != nil {
			return "0x0", types.NewRPCError(types.DefaultErrorCode, "failed to store offchain data")
		}

		return nil, nil
	})
	if err != nil {
		return "0x0", types.NewRPCError(types.DefaultErrorCode, "failed to store offchain data")
	}
	// Sign
	signedSequenceByMe, err := signedSequence.Sequence.Sign(d.privateKey)
	if err != nil {
		return "0x0", types.NewRPCError(types.DefaultErrorCode, "failed to sign")
	}
	// Return signature
	return signedSequenceByMe.Signature, nil
}
