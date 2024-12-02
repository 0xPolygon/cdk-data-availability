package synchronizer

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_getStartBlock(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		block   uint64
		wantErr bool
	}{
		{
			name: "GetLastProcessedBlock returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("GetLastProcessedBlock", mock.Anything, "L1").
					Return(uint64(0), testError)

				return mockDB
			},
			block:   0,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("GetLastProcessedBlock", mock.Anything, "L1").Return(uint64(5), nil)

				return mockDB
			},
			block: 4,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if block, err := getStartBlock(context.Background(), testDB, L1SyncTask); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.block, block)
			}
		})
	}
}

func Test_setStartBlock(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		block   uint64
		wantErr bool
	}{
		{
			name: "StoreLastProcessedBlock returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreLastProcessedBlock", mock.Anything, uint64(2), "L1").
					Return(testError)

				return mockDB
			},
			block:   2,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreLastProcessedBlock", mock.Anything, uint64(4), "L1").
					Return(nil)

				return mockDB
			},
			block: 4,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if err := setStartBlock(context.Background(), testDB, tt.block, L1SyncTask); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_storeMissingBatchKeys(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")
	testData := []types.BatchKey{
		{
			Number: 1,
			Hash:   common.HexToHash("0x01"),
		},
	}

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		keys    []types.BatchKey
		wantErr bool
	}{
		{
			name: "StoreMissingBatchKeys returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreMissingBatchKeys", mock.Anything, testData).Return(testError)

				return mockDB
			},
			keys:    testData,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreMissingBatchKeys", mock.Anything, testData).Return(nil)

				return mockDB
			},
			keys: testData,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if err := storeMissingBatchKeys(context.Background(), testDB, tt.keys); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_getMissingBatchKeys(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")
	testData := []types.BatchKey{
		{
			Number: 1,
			Hash:   common.HexToHash("0x01"),
		},
	}

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		keys    []types.BatchKey
		wantErr bool
	}{
		{
			name: "GetMissingBatchKeys returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("GetMissingBatchKeys", mock.Anything, uint(100)).
					Return(nil, testError)

				return mockDB
			},
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("GetMissingBatchKeys", mock.Anything, uint(100)).Return(testData, nil)

				return mockDB
			},
			keys: testData,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if keys, err := getMissingBatchKeys(context.Background(), testDB); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.keys, keys)
			}
		})
	}
}

func Test_deleteMissingBatchKeys(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")
	testData := []types.BatchKey{
		{
			Number: 1,
			Hash:   common.HexToHash("0x01"),
		},
	}

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		wantErr bool
	}{
		{
			name: "DeleteMissingBatchKeys returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("DeleteMissingBatchKeys", mock.Anything, testData).
					Return(testError)

				return mockDB
			},
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("DeleteMissingBatchKeys", mock.Anything, testData).
					Return(nil)

				return mockDB
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if err := deleteMissingBatchKeys(context.Background(), testDB, testData); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_storeOffchainData(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")
	testData := []types.OffChainData{
		{
			Key:   common.HexToHash("0x01"),
			Value: []byte("test data 1"),
		},
	}

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		data    []types.OffChainData
		wantErr bool
	}{
		{
			name: "StoreOffChainData returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreOffChainData", mock.Anything, testData).Return(testError)

				return mockDB
			},
			data:    testData,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreOffChainData", mock.Anything, testData).Return(nil)

				return mockDB
			},
			data: testData,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if err := storeOffchainData(context.Background(), testDB, tt.data); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
