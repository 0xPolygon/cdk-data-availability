package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestEndpoints_GetOffChainData(t *testing.T) {
	tests := []struct {
		name  string
		hash  types.ArgHash
		data  interface{}
		dbErr error
		err   error
	}{
		{
			name: "successfully got offchain data",
			hash: types.ArgHash{},
			data: types.ArgBytes("offchaindata"),
		},
		{
			name:  "db returns error",
			hash:  types.ArgHash{},
			data:  types.ArgBytes("offchaindata"),
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
				require.Equal(t, tt.data, got)
			}
		})
	}
}

func TestSyncEndpoints_ListOffChainData(t *testing.T) {
	tests := []struct {
		name   string
		hashes []types.ArgHash
		data   interface{}
		dbErr  error
		err    error
	}{
		{
			name:   "successfully got offchain data",
			hashes: []types.ArgHash{},
			data: map[common.Hash]types.ArgBytes{
				common.BytesToHash(nil): types.ArgBytes("offchaindata"),
			},
		},
		{
			name:   "db returns error",
			hashes: []types.ArgHash{},
			data: map[common.Hash]types.ArgBytes{
				common.BytesToHash(nil): types.ArgBytes("offchaindata"),
			},
			dbErr: errors.New("test error"),
			err:   errors.New("failed to list the requested data"),
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

			dbMock.On("ListOffChainData", context.Background(), keys).
				Return(tt.data, tt.dbErr)

			defer dbMock.AssertExpectations(t)

			z := &Endpoints{db: dbMock}

			got, err := z.ListOffChainData(tt.hashes)
			if tt.err != nil {
				require.Error(t, err)
				require.EqualError(t, tt.err, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.data, got)
			}
		})
	}
}
