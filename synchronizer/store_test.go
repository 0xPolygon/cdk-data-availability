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

func Test_storeUnresolvedBatchKeys(t *testing.T) {
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
			name: "StoreUnresolvedBatchKeys returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("StoreUnresolvedBatchKeys", mock.Anything, testData).Return(testError)

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

				mockDB.On("StoreUnresolvedBatchKeys", mock.Anything, testData).Return(nil)

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

			if err := storeUnresolvedBatchKeys(context.Background(), testDB, tt.keys); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_getUnresolvedBatchKeys(t *testing.T) {
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
			name: "GetUnresolvedBatchKeys returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("GetUnresolvedBatchKeys", mock.Anything, uint(100)).
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

				mockDB.On("GetUnresolvedBatchKeys", mock.Anything, uint(100)).Return(testData, nil)

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

			if keys, err := getUnresolvedBatchKeys(context.Background(), testDB); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.keys, keys)
			}
		})
	}
}

func Test_deleteUnresolvedBatchKeys(t *testing.T) {
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
			name: "DeleteUnresolvedBatchKeys returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()
				mockDB := mocks.NewDB(t)

				mockDB.On("DeleteUnresolvedBatchKeys", mock.Anything, testData).
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

				mockDB.On("DeleteUnresolvedBatchKeys", mock.Anything, testData).
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

			if err := deleteUnresolvedBatchKeys(context.Background(), testDB, testData); tt.wantErr {
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

func Test_detectOffchainDataGaps(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		gaps    map[uint64]uint64
		wantErr bool
	}{
		{
			name: "DetectOffchainDataGaps returns error",
			db: func(t *testing.T) db.DB {
				t.Helper()

				mockDB := mocks.NewDB(t)

				mockDB.On("DetectOffchainDataGaps", mock.Anything).Return(nil, testError)

				return mockDB
			},
			gaps:    nil,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				t.Helper()

				mockDB := mocks.NewDB(t)

				mockDB.On("DetectOffchainDataGaps", mock.Anything).Return(map[uint64]uint64{1: 3}, nil)

				return mockDB
			},
			gaps:    map[uint64]uint64{1: 3},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testDB := tt.db(t)

			if gaps, err := detectOffchainDataGaps(context.Background(), testDB); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.gaps, gaps)
			}
		})
	}
}
