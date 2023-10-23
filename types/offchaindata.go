package types

import "github.com/ethereum/go-ethereum/common"

// OffChainData represents some data that is not stored on chain and should be preserved
type OffChainData struct {
	Key   common.Hash
	Value []byte
}
