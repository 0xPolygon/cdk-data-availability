package types

import (
	"crypto/ecdsa"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	// ErrNonCanonicalSignature is returned when the signature is not canonical.
	ErrNonCanonicalSignature = errors.New("received non-canonical signature")
)

// Sign the hashToSIgn with the given privateKey.
func Sign(privateKey *ecdsa.PrivateKey, hashToSign []byte) ([]byte, error) {
	sig, err := crypto.Sign(hashToSign, privateKey)
	if err != nil {
		return nil, err
	}

	if strings.ToUpper(common.Bytes2Hex(sig[32:64])) > "7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0" {
		return nil, ErrNonCanonicalSignature
	}

	sig[64] += 27

	return sig, nil
}
