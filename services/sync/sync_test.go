package sync

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestEndpoints_GetOffChainData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		hash  types.ArgHash
		data  *types.OffChainData
		dbErr error
		err   error
	}{
		{
			name: "successfully got offchain data",
			hash: types.ArgHash{},
			data: &types.OffChainData{
				Key:      common.Hash{},
				Value:    types.ArgBytes("offchaindata"),
				BatchNum: 0,
			},
		},
		{
			name: "db returns error",
			hash: types.ArgHash{},
			data: &types.OffChainData{
				Key:      common.Hash{},
				Value:    types.ArgBytes("offchaindata"),
				BatchNum: 0,
			},
			dbErr: errors.New("test error"),
			err:   errors.New("failed to get the requested data"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbMock := mocks.NewDB(t)

			dbMock.On("GetOffChainData", context.Background(), tt.hash.Hash()).
				Return(tt.data, tt.dbErr)

			defer dbMock.AssertExpectations(t)

			z := &Endpoints{db: dbMock}

			got, err := z.GetOffChainData(tt.hash)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, types.ArgBytes(tt.data.Value), got)
			}
		})
	}
}

func TestSyncEndpoints_ListOffChainData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		hashes []types.ArgHash
		data   []types.OffChainData
		dbErr  error
		err    error
	}{
		{
			name:   "successfully got offchain data",
			hashes: generateRandomHashes(t, 1),
			data: []types.OffChainData{{
				Key:      common.BytesToHash(nil),
				Value:    types.ArgBytes("offchaindata"),
				BatchNum: 0,
			}},
		},
		{
			name:   "db returns error",
			hashes: []types.ArgHash{},
			data: []types.OffChainData{{
				Key:      common.BytesToHash(nil),
				Value:    types.ArgBytes("offchaindata"),
				BatchNum: 0,
			}},
			dbErr: errors.New("test error"),
			err:   errors.New("failed to list the requested data"),
		},
		{
			name:   "too many hashes requested",
			hashes: generateRandomHashes(t, maxListHashes+1),
			err:    errors.New("too many hashes requested"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbMock := mocks.NewDB(t)

			keys := make([]common.Hash, len(tt.hashes))
			for i, hash := range tt.hashes {
				keys[i] = hash.Hash()
			}

			if tt.data != nil {
				dbMock.On("ListOffChainData", context.Background(), keys).
					Return(tt.data, tt.dbErr)

				defer dbMock.AssertExpectations(t)
			}

			z := &Endpoints{db: dbMock}

			got, err := z.ListOffChainData(tt.hashes)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)

				listMap := make(map[common.Hash]types.ArgBytes)
				for _, data := range tt.data {
					listMap[data.Key] = data.Value
				}

				require.Equal(t, listMap, got)
			}
		})
	}
}

func generateRandomHashes(t *testing.T, numOfHashes int) []types.ArgHash {
	t.Helper()

	hashes := make([]types.ArgHash, numOfHashes)
	for i := 0; i < numOfHashes; i++ {
		hashes[i] = types.ArgHash(generateRandomHash(t))
	}

	return hashes
}

func generateRandomHash(t *testing.T) common.Hash {
	t.Helper()

	randomData := make([]byte, 32)

	_, err := rand.Read(randomData)
	require.NoError(t, err)

	return crypto.Keccak256Hash(randomData)
}
