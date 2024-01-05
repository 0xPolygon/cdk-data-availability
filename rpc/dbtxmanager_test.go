package rpc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0xPolygon/cdk-data-availability/db"
	"github.com/0xPolygon/cdk-data-availability/mocks"
	"github.com/0xPolygon/cdk-data-availability/rpc"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDBTxManager_NewDbTxScope(t *testing.T) {
	testTable := []struct {
		name      string
		db        func(db *mocks.DB) *mocks.Tx
		scopedFn  rpc.DBTxScopedFn
		expected  interface{}
		returnErr error
	}{
		{
			name: "successfully executed scoped function",
			db: func(dbMock *mocks.DB) *mocks.Tx {
				txMock := mocks.NewTx(t)
				dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil)
				txMock.On("Commit").Return(nil)
				return txMock
			},
			scopedFn: func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
				return 123, nil
			},
			expected: 123,
		},
		{
			name: "BeginStateTransaction returns error",
			db: func(dbMock *mocks.DB) *mocks.Tx {
				dbMock.On("BeginStateTransaction", mock.Anything).
					Return(nil, errors.New("test"))
				return nil
			},
			scopedFn: func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
				t.Fatal("unexpected scopedFn execution")
				return nil, nil
			},
			returnErr: errors.New("failed to connect to the state"),
		},
		{
			name: "scopedFn returns error",
			db: func(dbMock *mocks.DB) *mocks.Tx {
				txMock := mocks.NewTx(t)
				dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil)
				txMock.On("Rollback").Return(nil)
				return txMock
			},
			scopedFn: func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
				return nil, rpc.NewRPCError(1, "test")
			},
			returnErr: errors.New("test"),
		},
		{
			name: "Rollback returns error",
			db: func(dbMock *mocks.DB) *mocks.Tx {
				txMock := mocks.NewTx(t)
				dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil)
				txMock.On("Rollback").Return(errors.New("test"))
				return txMock
			},
			scopedFn: func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
				return nil, rpc.NewRPCError(1, "test")
			},
			returnErr: errors.New("failed to rollback db transaction"),
		},
		{
			name: "Commit returns error",
			db: func(dbMock *mocks.DB) *mocks.Tx {
				txMock := mocks.NewTx(t)
				dbMock.On("BeginStateTransaction", mock.Anything).Return(txMock, nil)
				txMock.On("Commit").Return(errors.New("test"))
				return txMock
			},
			scopedFn: func(ctx context.Context, dbTx db.Tx) (interface{}, rpc.Error) {
				return 123, nil
			},
			returnErr: errors.New("failed to commit db transaction"),
		},
	}

	for _, tt := range testTable {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbMock := mocks.NewDB(t)
			txMock := tt.db(dbMock)

			mgr := rpc.DBTxManager{}

			got, err := mgr.NewDbTxScope(dbMock, tt.scopedFn)
			if tt.returnErr != nil {
				require.Equal(t, tt.returnErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, got)
			}

			dbMock.AssertExpectations(t)

			if txMock != nil {
				txMock.AssertExpectations(t)
			}
		})
	}
}
