package dac

import (
	"fmt"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/mocks"
)

func TestStatusEndpoints_GetStatus(t *testing.T) {
	tests := []struct {
		name             string
		dbErr            error
		backfillProgress uint64
		expectedUptime   string
		expectedVersion  string
		expectedKeyCount uint64
		expectedError    error
	}{
		{
			name:             "successfully get status",
			backfillProgress: 1000,
			expectedUptime:   "1h30m",
			expectedVersion:  "1.0.0",
			expectedKeyCount: 100,
			dbErr:            nil,
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
			fmt.Println(dbMock)

			k := status{}
			dbMock.On("GetStatus").
				Return(k, tt.dbErr)

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
