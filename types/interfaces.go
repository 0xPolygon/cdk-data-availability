package types

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

// SignedSequenceInterface is the interface that defines the methods that a signed sequence must implement
type SignedSequenceInterface interface {
	Signer() (common.Address, error)
	OffChainData() []OffChainData
	Sign(privateKey *ecdsa.PrivateKey) (ArgBytes, error)
	SetSignature([]byte)
	GetSignature() []byte
}
