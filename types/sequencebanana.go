package types

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/keccak256"
)

// Batch represents the batch data that the sequencer will send to L1
type Batch struct {
	L2Data            ArgBytes       `json:"L2Data"`
	ForcedGER         common.Hash    `json:"forcedGlobalExitRoot"`
	ForcedTimestamp   ArgUint64      `json:"forcedTimestamp"`
	Coinbase          common.Address `json:"coinbase"`
	ForcedBlockHashL1 common.Hash    `json:"forcedBlockHashL1"`
}

// SequenceBanana represents the data that the sequencer will send to L1
// and other metadata needed to build the accumulated input hash aka accInputHash
type SequenceBanana struct {
	Batches              []Batch     `json:"batches"`
	OldAccInputHash      common.Hash `json:"oldAccInputhash"`
	L1InfoRoot           common.Hash `json:"l1InfoRoot"`
	MaxSequenceTimestamp ArgUint64   `json:"maxSequenceTimestamp"`
}

// HashToSign returns the accumulated input hash of the sequence.
// Note that this is equivalent to what happens on the smart contract
func (s *SequenceBanana) HashToSign() []byte {
	v1 := s.OldAccInputHash.Bytes()
	for _, b := range s.Batches {
		v2 := b.L2Data
		var v3, v4 []byte
		if b.ForcedTimestamp > 0 {
			v3 = b.ForcedGER.Bytes()
			v4 = big.NewInt(0).SetUint64(uint64(b.ForcedTimestamp)).Bytes()
		} else {
			v3 = s.L1InfoRoot.Bytes()
			v4 = big.NewInt(0).SetUint64(uint64(s.MaxSequenceTimestamp)).Bytes()
		}
		v5 := b.Coinbase.Bytes()
		v6 := b.ForcedBlockHashL1.Bytes()

		// Add 0s to make values 32 bytes long
		for len(v1) < 32 {
			v1 = append([]byte{0}, v1...)
		}
		v2 = keccak256.Hash(v2)
		for len(v3) < 32 {
			v3 = append([]byte{0}, v3...)
		}
		for len(v4) < 8 {
			v4 = append([]byte{0}, v4...)
		}
		for len(v5) < 20 {
			v5 = append([]byte{0}, v5...)
		}
		for len(v6) < 32 {
			v6 = append([]byte{0}, v6...)
		}
		v1 = keccak256.Hash(v1, v2, v3, v4, v5, v6)
	}

	return v1
}

// Sign returns a signed sequence by the private key.
// Note that what's being signed is the accumulated input hash
func (s *SequenceBanana) Sign(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hashToSign := s.HashToSign()
	return Sign(privateKey, hashToSign)
}

// OffChainData returns the data that needs to be stored off chain from a given sequence
func (s *SequenceBanana) OffChainData() []OffChainData {
	od := []OffChainData{}
	for _, b := range s.Batches {
		od = append(od, OffChainData{
			Key:   crypto.Keccak256Hash(b.L2Data),
			Value: b.L2Data,
		})
	}
	return od
}

// SignedSequenceBanana is a sequence but signed
type SignedSequenceBanana struct {
	Sequence  SequenceBanana `json:"sequence"`
	Signature ArgBytes       `json:"signature"`
}

// Signer returns the address of the signer
func (s *SignedSequenceBanana) Signer() (common.Address, error) {
	if len(s.Signature) != signatureLen {
		return common.Address{}, errors.New("invalid signature")
	}
	sig := make([]byte, signatureLen)
	copy(sig, s.Signature)
	sig[64] -= 27
	pubKey, err := crypto.SigToPub(s.Sequence.HashToSign(), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubKey), nil
}

// OffChainData returns the data to be stored of the sequence
func (s *SignedSequenceBanana) OffChainData() []OffChainData {
	return s.Sequence.OffChainData()
}

// Sign signs the sequence using the privateKey
func (s *SignedSequenceBanana) Sign(privateKey *ecdsa.PrivateKey) (ArgBytes, error) {
	return s.Sequence.Sign(privateKey)
}

// SetSignature set signature
func (s *SignedSequenceBanana) SetSignature(sign []byte) {
	s.Signature = sign
}

// GetSignature returns signature
func (s *SignedSequenceBanana) GetSignature() []byte {
	return s.Signature
}
