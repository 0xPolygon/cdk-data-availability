package status

import (
	"testing"

	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEndpoints_GetStatus(t *testing.T) {
	tests := []struct {
		name                       string
		getOffchainDataRowCountErr error
		getLastProcessedBlockErr   error
		backfillProgress           uint64
		expectedUptime             string
		expectedVersion            string
		expectedKeyCount           uint64
		expectedError              error
	}{
		{
			name:                       "successfully get status",
			backfillProgress:           1000,
			expectedVersion:            "v1.0.0",
			expectedKeyCount:           100,
			getOffchainDataRowCountErr: nil,
			getLastProcessedBlockErr:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbMock := mocks.NewDB(t)

			dbMock.On("GetOffchainDataRowCount", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					require.Len(t, args, 2)

					rowCount := args[1].(*uint64)
					*rowCount = tt.expectedKeyCount
				}).
				Return(tt.getLastProcessedBlockErr)

			dbMock.On("GetLastProcessedBlock", mock.Anything, mock.Anything).
				Return(tt.backfillProgress, tt.getLastProcessedBlockErr)

			statusEndpoints := NewEndpoints(dbMock)

			actual, err := statusEndpoints.GetStatus()

			if tt.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)

				require.NotEmpty(t, actual.(types.DACStatus).Uptime)
				require.Equal(t, "v0.1.0", actual.(types.DACStatus).Version)
				require.Equal(t, tt.expectedKeyCount, actual.(types.DACStatus).KeyCount)
				require.Equal(t, tt.backfillProgress, actual.(types.DACStatus).BackfillProgress)
			}
		})
	}
}
