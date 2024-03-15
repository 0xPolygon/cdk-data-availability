package dac

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/0xPolygon/cdk-data-availability/mocks"
)

func TestStatusEndpoints_GetStatus(t *testing.T) {
	tests := []struct {
		name                     string
		getRowCountErr           error
		getLastProcessedBlockErr error
		backfillProgress         uint64
		expectedUptime           string
		expectedVersion          string
		expectedKeyCount         uint64
		expectedError            error
	}{
		{
			name:                     "successfully get status",
			backfillProgress:         1000,
			expectedVersion:          "v1.0.0",
			expectedKeyCount:         100,
			getRowCountErr:           nil,
			getLastProcessedBlockErr: nil,
		},
		// {
		// 	name:          "database error",
		// 	dbErr:         errors.New("test database error"),
		// 	expectedError: errors.New("failed to get the key count from the offchain_data table: test database error"),
		// },
		// {
		// 	name:          "backfill progress error",
		// 	dbErr:         nil,
		// 	expectedError: errors.New("failed to get last block processed by the synchronizer: test backfill progress error"),
		// },
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbMock := mocks.NewDB(t)

			dbMock.On("GetRowCount", mock.Anything, mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					require.Len(t, args, 3)

					rowCount := args[1].(*uint64)
					*rowCount = tt.expectedKeyCount
				}).
				Return(tt.getRowCountErr)

			dbMock.On("GetLastProcessedBlock", mock.Anything, mock.Anything).
				Return(tt.backfillProgress, tt.getLastProcessedBlockErr)

			dacEndpoints := NewDacEndpoints(dbMock)

			actual, err := dacEndpoints.GetStatus()

			if tt.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)

				require.NotEmpty(t, actual.(Status).Uptime)
				require.Equal(t, "v0.1.0", actual.(Status).Version)
				require.Equal(t, tt.expectedKeyCount, actual.(Status).KeyCount)
				require.Equal(t, tt.backfillProgress, actual.(Status).BackfillProgress)
			}

			// dbMock.On("GetRowCount", "SELECT COUNT(*) FROM data_node.offchain_data;", (*uint64)(nil), context.Background()).
			// 	Return(tt.dbErr)
			// if tt.dbErr == nil {
			// 	dbMock.On("",)
			// }

			// if tt.dbErr == nil {
			// 	dbMock.On("GetLastProcessedBlock", context.Background(), synchronizer.L1SyncTask).
			// 		Return(tt.backfillProgress, tt.dbErr)
			// }

			// defer dbMock.AssertExpectations(t)

			// startTime := time.Now().Add(-time.Hour * 1) // Subtracting 1 hour for testing uptime
			// statusEndpoint := &StatusEndpoints{
			// 	db:        dbMock,
			// 	startTime: startTime,
			// }

			// cur_status, err := statusEndpoint.GetStatus()

			// if tt.expectedError != nil {
			// 	require.Error(t, err)
			// 	require.EqualError(t, err, tt.expectedError.Error())
			// } else {
			// 	require.NoError(t, err)

			// 	expectedStatus := status{
			// 		uptime:           tt.expectedUptime,
			// 		version:          "1.0.0", // Assuming a fixed version for testing
			// 		keyCount:         tt.expectedKeyCount,
			// 		backfillProgress: tt.backfillProgress,
			// 	}

			// 	require.Equal(t, expectedStatus, cur_status)
			// }
		})
	}
}
