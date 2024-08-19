package datacom

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/log"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/0xPolygon/cdk-data-availability/sequencer"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
)

// APIDATACOM is the namespace of the datacom service
const APIDATACOM = "datacom"

// Endpoints contains implementations for the "datacom" RPC endpoints
type Endpoints struct {
	db               db.DB
	privateKey       *ecdsa.PrivateKey
	sequencerTracker *sequencer.Tracker
}

// NewEndpoints returns Endpoints
func NewEndpoints(db db.DB, pk *ecdsa.PrivateKey, st *sequencer.Tracker) *Endpoints {
	return &Endpoints{
		db:               db,
		privateKey:       pk,
		sequencerTracker: st,
	}
}

// SignSequence generates the concatenation of hashes of the batch data of the sequence and sign it.
// After storing the data that will be sent hashed to the contract, it returns the signature.
// This endpoint is only accessible to the sequencer
func (d *Endpoints) SignSequence(signedSequence types.SignedSequence) (interface{}, rpc.Error) {
	return d.signSequence(&signedSequence)
}

// SignSequenceBanana generates the accumulated input hash aka accInputHash of the sequence and sign it.
// After storing the data that will be sent hashed to the contract, it returns the signature.
// This endpoint is only accessible to the sequencer
func (d *Endpoints) SignSequenceBanana(signedSequence types.SignedSequenceBanana) (interface{}, rpc.Error) {
	log.Debugf("signing sequence, hash to sign: %s", common.BytesToHash(signedSequence.Sequence.HashToSign()))
	return d.signSequence(&signedSequence)
}

func (d *Endpoints) signSequence(signedSequence types.SignedSequenceInterface) (interface{}, rpc.Error) {
	// Verify that the request comes from the sequencer
	sender, err := signedSequence.Signer()
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, "failed to verify sender")
	}

	if sender != d.sequencerTracker.GetAddr() {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, "unauthorized")
	}

	// Store off-chain data by hash (hash(L2Data): L2Data)
	if err = d.db.StoreOffChainData(context.Background(), signedSequence.OffChainData()); err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode,
			fmt.Errorf("failed to store offchain data. Error: %w", err).Error())
	}

	// Sign
	signature, err := signedSequence.Sign(d.privateKey)
	if err != nil {
		return "0x0", rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Errorf("failed to sign. Error: %w", err).Error())
	}
	// Return signature
	return signature, nil
}
