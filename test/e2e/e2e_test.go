package e2e

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/client"
	"github.com/0xPolygon/cdk-data-availability/config"
	cfgTypes "github.com/0xPolygon/cdk-data-availability/config/types"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T) (*testClient, *ecdsa.PrivateKey) {
	const (
		url         = "http://localhost:8444"
		memberAddr  = "0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
		privKeyPath = "../config/sequencer.keystore"
		privKeyPass = "testonly"
	)
	pk, err := config.NewKeyFromKeystore(cfgTypes.KeystoreFileConfig{
		Path:     privKeyPath,
		Password: privKeyPass,
	})
	require.NoError(t, err)
	return newTestClient(url, common.HexToAddress(memberAddr)), pk
}

func TestSignSequence(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tc, pk := initTest(t)
	type testSequences struct {
		name        string
		sequence    types.SignedSequence
		expectedErr error
	}
	expectedSequence := types.Sequence{
		common.Hex2Bytes("274567245673256275642756243560234572347657236520"),
		common.Hex2Bytes("7863456782345678923456789354"),
	}

	unexpectedSenderPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	unexpectedSenderSignature, err := expectedSequence.Sign(unexpectedSenderPrivKey)
	require.NoError(t, err)
	tSequences := []testSequences{
		{
			name: "invalid_signature",
			sequence: types.SignedSequence{
				Sequence:  types.Sequence{},
				Signature: common.Hex2Bytes("f00"),
			},
			expectedErr: errors.New("-32000 failed to verify sender"),
		},
		{
			name: "signature_not_from_sender",
			sequence: types.SignedSequence{
				Sequence:  expectedSequence,
				Signature: unexpectedSenderSignature,
			},
			expectedErr: errors.New("-32000 unauthorized"),
		},
		{
			name:        "empty_batch",
			sequence:    types.SignedSequence{},
			expectedErr: nil,
		},
		{
			name: "success",
			sequence: types.SignedSequence{
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
				ts.sequence = types.SignedSequence{
					Sequence:  ts.sequence.Sequence,
					Signature: signature,
				}
			}
			tc.signSequence(t, &ts.sequence, ts.expectedErr)
		})
	}
}

type testClient struct {
	client        client.Client
	dacMemberAddr common.Address
}

func newTestClient(url string, addr common.Address) *testClient {
	return &testClient{
		client:        client.New(url),
		dacMemberAddr: addr,
	}
}

func (tc *testClient) signSequence(t *testing.T, expected *types.SignedSequence, expectedErr error) {
	if signature, err := tc.client.SignSequence(context.Background(), *expected); err != nil {
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
