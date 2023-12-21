package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/stretchr/testify/require"
)

func TestSyncEndpoints_GetOffChainData(t *testing.T) {
	tests := []struct {
		name  string
		hash  types.ArgHash
		data  interface{}
		dbErr error
		txErr error
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
		{
			name:  "tx returns error",
			hash:  types.ArgHash{},
			data:  types.ArgBytes("offchaindata"),
			txErr: errors.New("test error"),
			err:   errors.New("failed to connect to the state"),
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			txMock := mocks.NewTx(t)

			dbMock := mocks.NewDB(t)
			dbMock.On("BeginStateTransaction", context.Background()).
				Return(txMock, tt.txErr)
			if tt.txErr == nil {
				dbMock.On("GetOffChainData", context.Background(), tt.hash.Hash(), txMock).
					Return(tt.data, tt.dbErr)

				if tt.err != nil {
					txMock.On("Rollback").
						Return(nil)
				} else {
					txMock.On("Commit").
						Return(nil)
				}
			}

			defer txMock.AssertExpectations(t)
			defer dbMock.AssertExpectations(t)

			z := &SyncEndpoints{db: dbMock}

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
