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
		Batches: []types.Batch{
			{
				Number:         3,
				GlobalExitRoot: common.HexToHash("0x678343456734678"),
				Timestamp:      3457834,
				Coinbase:       common.HexToAddress("0x345678934t567889137"),
				L2Data:         common.Hex2Bytes("274567245673256275642756243560234572347657236520"),
			},
			{
				Number:         4,
				GlobalExitRoot: common.HexToHash("0x34678345678345789534678912345678"),
				Timestamp:      78907890,
				Coinbase:       common.HexToAddress("0x3457234672345789234567"),
				L2Data:         common.Hex2Bytes("7863456782345678923456789354"),
			},
		},
	}

	unexpectedSenderPrivKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	unexpectedSenderSignedSequence, err := expectedSequence.Sign(unexpectedSenderPrivKey)
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
				Signature: unexpectedSenderSignedSequence.Signature,
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
				signedBatch, err := ts.sequence.Sequence.Sign(pk)
				require.NoError(t, err)
				ts.sequence = *signedBatch
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
		client:        *client.New(url),
		dacMemberAddr: addr,
	}
}

func (tc *testClient) signSequence(t *testing.T, expected *types.SignedSequence, expectedErr error) {
	if signature, err := tc.client.SignSequence(*expected); err != nil {
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
	}
}
