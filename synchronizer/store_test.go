package synchronizer

import (
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_setStartBlock(t *testing.T) {
	testError := errors.New("test error")

	tests := []struct {
		name    string
		db      func(t *testing.T) db.DB
		block   uint64
		wantErr bool
	}{
		{
			name: "BeginStateTransaction returns error",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockDB.On("BeginStateTransaction", mock.Anything).
					Return(nil, testError)

				return mockDB
			},
			block:   1,
			wantErr: true,
		},
		{
			name: "StoreLastProcessedBlock returns error",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockTx := mocks.NewTx(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(mockTx, nil)

				mockDB.On("StoreLastProcessedBlock", mock.Anything, "L1", uint64(2), mockTx).
					Return(testError)

				return mockDB
			},
			block:   2,
			wantErr: true,
		},
		{
			name: "Commit returns error",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockTx := mocks.NewTx(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(mockTx, nil)

				mockDB.On("StoreLastProcessedBlock", mock.Anything, "L1", uint64(3), mockTx).
					Return(nil)

				mockTx.On("Commit").
					Return(testError)

				return mockDB
			},
			block:   3,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockTx := mocks.NewTx(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(mockTx, nil)

				mockDB.On("StoreLastProcessedBlock", mock.Anything, "L1", uint64(4), mockTx).
					Return(nil)

				mockTx.On("Commit").
					Return(nil)

				return mockDB
			},
			block: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := tt.db(t)

			if err := setStartBlock(testDB, tt.block); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_exists(t *testing.T) {
	tests := []struct {
		name string
		db   func(t *testing.T) db.DB
		key  common.Hash
		want bool
	}{
		{
			name: "Exists returns true",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockDB.On("Exists", mock.Anything, common.HexToHash("0x01")).
					Return(true)

				return mockDB
			},
			key:  common.HexToHash("0x01"),
			want: true,
		},
		{
			name: "Exists returns false",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockDB.On("Exists", mock.Anything, common.HexToHash("0x02")).
					Return(false)

				return mockDB
			},
			key:  common.HexToHash("0x02"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := tt.db(t)

			got := exists(testDB, tt.key)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_store(t *testing.T) {
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
			name: "BeginStateTransaction returns error",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(nil, testError)

				return mockDB
			},
			data:    testData,
			wantErr: true,
		},
		{
			name: "StoreOffChainData returns error",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockTx := mocks.NewTx(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(mockTx, nil)

				mockDB.On("StoreOffChainData", mock.Anything, testData, mockTx).Return(testError)

				mockTx.On("Rollback").Return(nil)

				return mockDB
			},
			data:    testData,
			wantErr: true,
		},
		{
			name: "Commit returns error",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockTx := mocks.NewTx(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(mockTx, nil)

				mockDB.On("StoreOffChainData", mock.Anything, testData, mockTx).Return(nil)

				mockTx.On("Commit").Return(testError)

				return mockDB
			},
			data:    testData,
			wantErr: true,
		},
		{
			name: "all good",
			db: func(t *testing.T) db.DB {
				mockDB := mocks.NewDB(t)

				mockTx := mocks.NewTx(t)

				mockDB.On("BeginStateTransaction", mock.Anything).Return(mockTx, nil)

				mockDB.On("StoreOffChainData", mock.Anything, testData, mockTx).Return(nil)

				mockTx.On("Commit").Return(nil)

				return mockDB
			},
			data: testData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB := tt.db(t)

			if err := store(testDB, tt.data); tt.wantErr {
				require.ErrorIs(t, err, testError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
