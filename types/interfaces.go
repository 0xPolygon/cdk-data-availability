package types

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

type SignedSequenceInterface interface {
	Signer() (common.Address, error)
	OffChainData() []OffChainData
	Sign(privateKey *ecdsa.PrivateKey) (ArgBytes, error)
}
