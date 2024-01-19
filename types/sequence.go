package types

import (
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	solsha3 "github.com/miguelmota/go-solidity-sha3"
)

const (
	signatureLen = 65
)

// Sequence represents the data that the sequencer will send to L1
// and other metadata needed to build the accumulated input hash aka accInputHash
type Sequence []ArgBytes

// HashToSign returns the accumulated input hash of the sequence.
// Note that this is equivalent to what happens on the smart contract
func (s *Sequence) HashToSign() []byte {
	currentHash := common.Hash{}.Bytes()
	for _, batchData := range ([]ArgBytes)(*s) {
		types := []string{
			"bytes32",
			"bytes32",
		}
		values := []interface{}{
			currentHash,
			crypto.Keccak256(batchData),
		}
		currentHash = solsha3.SoliditySHA3(types, values)
	}
	return currentHash
}

// Sign returns a signed sequence by the private key.
// Note that what's being signed is the accumulated input hash
func (s *Sequence) Sign(privateKey *ecdsa.PrivateKey) (*SignedSequence, error) {
	hashToSign := s.HashToSign()
	sig, err := crypto.Sign(hashToSign, privateKey)
	if err != nil {
		return nil, err
	}

	rBytes := sig[:32]
	sBytes := sig[32:64]
	vByte := sig[64]

	if strings.ToUpper(common.Bytes2Hex(sBytes)) > "7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF5D576E7357A4501DDFE92F46681B20A0" {
		magicNumber := common.Hex2Bytes("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141")
		sBig := big.NewInt(0).SetBytes(sBytes)
		magicBig := big.NewInt(0).SetBytes(magicNumber)
		s1 := magicBig.Sub(magicBig, sBig)
		sBytes = s1.Bytes()
		if vByte == 0 {
			vByte = 1
		} else {
			vByte = 0
		}
	}
	vByte += 27

	actualSignature := []byte{}
	actualSignature = append(actualSignature, rBytes...)
	actualSignature = append(actualSignature, sBytes...)
	actualSignature = append(actualSignature, vByte)

	return &SignedSequence{
		Sequence:  *s,
		Signature: actualSignature,
	}, nil
}

// OffChainData returns the data that needs to be stored off chain from a given sequence
func (s *Sequence) OffChainData() []OffChainData {
	od := []OffChainData{}
	for _, batchData := range ([]ArgBytes)(*s) {
		od = append(od, OffChainData{
			Key:   crypto.Keccak256Hash(batchData),
			Value: batchData,
		})
	}
	return od
}

// SignedSequence is a sequence but signed
type SignedSequence struct {
	Sequence  Sequence `json:"sequence"`
	Signature ArgBytes `json:"signature"`
}

// Signer returns the address of the signer
func (s *SignedSequence) Signer() (common.Address, error) {
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
