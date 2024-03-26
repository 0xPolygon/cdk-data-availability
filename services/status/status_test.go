package status

import (
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEndpoints_GetStatus(t *testing.T) {
	tests := []struct {
		name                     string
		countOffchainData        uint64
		countOffchainDataErr     error
		getLastProcessedBlock    uint64
		getLastProcessedBlockErr error
		expectedError            error
	}{
		{
			name:                  "successfully got status",
			countOffchainData:     1,
			getLastProcessedBlock: 2,
		},
		{
			name:                  "failed to count offchain data",
			countOffchainDataErr:  errors.New("test error"),
			getLastProcessedBlock: 2,
		},
		{
			name:                     "failed to count offchain data and last processed block",
			countOffchainDataErr:     errors.New("test error"),
			getLastProcessedBlockErr: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbMock := mocks.NewDB(t)

			dbMock.On("CountOffchainData", mock.Anything).
				Return(tt.countOffchainData, tt.countOffchainDataErr)

			dbMock.On("GetLastProcessedBlock", mock.Anything, mock.Anything).
				Return(tt.getLastProcessedBlock, tt.getLastProcessedBlockErr)

			statusEndpoints := NewEndpoints(dbMock)

			actual, err := statusEndpoints.GetStatus()

			if tt.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tt.expectedError.Error())
			} else {
				require.NoError(t, err)

				require.NotEmpty(t, actual.(types.DACStatus).Uptime)
				require.Equal(t, "v0.1.0", actual.(types.DACStatus).Version)
				require.Equal(t, tt.countOffchainData, actual.(types.DACStatus).KeyCount)
				require.Equal(t, tt.getLastProcessedBlock, actual.(types.DACStatus).BackfillProgress)
			}
		})
	}
}
