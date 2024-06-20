package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignSequenceBanana(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tc, pk := initTest(t)
	type testSequences struct {
		name        string
		sequence    types.SignedSequenceBanana
		expectedErr error
	}
	expectedSequence := types.SequenceBanana{
		Batches: []types.Batch{
			{
				L2Data:            common.FromHex("723473475757adadaddada"),
				ForcedGER:         common.Hash{},
				ForcedTimestamp:   0,
				Coinbase:          common.HexToAddress("aabbccddee"),
				ForcedBlockHashL1: common.Hash{},
			},
			{
				L2Data:            common.FromHex("723473475757adadaddada723473475757adadaddada"),
				ForcedGER:         common.Hash{},
				ForcedTimestamp:   0,
				Coinbase:          common.HexToAddress("aabbccddee"),
				ForcedBlockHashL1: common.Hash{},
			},
		},
		OldAccInputHash:      common.HexToHash("abcdef0987654321"),
		L1InfoRoot:           common.HexToHash("ffddeeaabb09876"),
		MaxSequenceTimestamp: 78945,
	}

	unexpectedSenderPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	unexpectedSenderSignature, err := expectedSequence.Sign(unexpectedSenderPrivKey)
	require.NoError(t, err)
	tSequences := []testSequences{
		{
			name: "invalid_signature",
			sequence: types.SignedSequenceBanana{
				Sequence:  types.SequenceBanana{},
				Signature: common.Hex2Bytes("f00"),
			},
			expectedErr: errors.New("-32000 failed to verify sender"),
		},
		{
			name: "signature_not_from_sender",
			sequence: types.SignedSequenceBanana{
				Sequence:  expectedSequence,
				Signature: unexpectedSenderSignature,
			},
			expectedErr: errors.New("-32000 unauthorized"),
		},
		{
			name:        "empty_batch",
			sequence:    types.SignedSequenceBanana{},
			expectedErr: nil,
		},
		{
			name: "success",
			sequence: types.SignedSequenceBanana{
				Sequence:  expectedSequence,
				Signature: nil,
			},
			expectedErr: nil,
		},
	}
	for _, ts := range tSequences {
		t.Run(ts.name, func(t *testing.T) {
			if ts.sequence.Signature == nil {
				signature, err := ts.sequence.Sequence.Sign(pk)
				require.NoError(t, err)
				ts.sequence = types.SignedSequenceBanana{
					Sequence:  ts.sequence.Sequence,
					Signature: signature,
				}
			}
			tc.signSequenceBanana(t, &ts.sequence, ts.expectedErr)
		})
	}
}

func (tc *testClient) signSequenceBanana(t *testing.T, expected *types.SignedSequenceBanana, expectedErr error) {
	if signature, err := tc.client.SignSequenceBanana(context.Background(), *expected); err != nil {
		assert.Equal(t, expectedErr.Error(), err.Error())
	} else {
		// Verify signature
		require.NoError(t, expectedErr)
		expected.Signature = signature
		actualAddr, err := expected.Signer()
		require.NoError(t, err)
		assert.Equal(t, tc.dacMemberAddr, actualAddr)

		// Check that offchain data has been stored
		expectedOffchainData := expected.Sequence.OffChainData()
		for _, od := range expectedOffchainData {
			actualData, err := tc.client.GetOffChainData(
				context.Background(),
				od.Key,
			)
			require.NoError(t, err)
			assert.Equal(t, od.Value, actualData)
		}

		hashes := make([]common.Hash, len(expectedOffchainData))
		for i, od := range expectedOffchainData {
			hashes[i] = od.Key
		}

		actualData, err := tc.client.ListOffChainData(context.Background(), hashes)
		require.NoError(t, err)

		for _, od := range expectedOffchainData {
			assert.Equal(t, od.Value, actualData[od.Key])
		}
	}
}
