// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	context "context"

	common "github.com/ethereum/go-ethereum/common"

	db "github.com/0xPolygon/cdk-data-availability/db"

	mock "github.com/stretchr/testify/mock"

	sqlx "github.com/jmoiron/sqlx"

	types "github.com/0xPolygon/cdk-data-availability/types"
)

// DB is an autogenerated mock type for the DB type
type DB struct {
	mock.Mock
}

// BeginStateTransaction provides a mock function with given fields: ctx
func (_m *DB) BeginStateTransaction(ctx context.Context) (db.Tx, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for BeginStateTransaction")
	}

	var r0 db.Tx
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (db.Tx, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) db.Tx); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(db.Tx)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Exists provides a mock function with given fields: ctx, key
func (_m *DB) Exists(ctx context.Context, key common.Hash) bool {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash) bool); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetLastProcessedBlock provides a mock function with given fields: ctx, task
func (_m *DB) GetLastProcessedBlock(ctx context.Context, task string) (uint64, error) {
	ret := _m.Called(ctx, task)

	if len(ret) == 0 {
		panic("no return value specified for GetLastProcessedBlock")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (uint64, error)); ok {
		return rf(ctx, task)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) uint64); ok {
		r0 = rf(ctx, task)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, task)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOffChainData provides a mock function with given fields: ctx, key, dbTx
func (_m *DB) GetOffChainData(ctx context.Context, key common.Hash, dbTx sqlx.QueryerContext) (types.ArgBytes, error) {
	ret := _m.Called(ctx, key, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for GetOffChainData")
	}

	var r0 types.ArgBytes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash, sqlx.QueryerContext) (types.ArgBytes, error)); ok {
		return rf(ctx, key, dbTx)
	}
	if rf, ok := ret.Get(0).(func(context.Context, common.Hash, sqlx.QueryerContext) types.ArgBytes); ok {
		r0 = rf(ctx, key, dbTx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.ArgBytes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, common.Hash, sqlx.QueryerContext) error); ok {
		r1 = rf(ctx, key, dbTx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StoreLastProcessedBlock provides a mock function with given fields: ctx, task, block, dbTx
func (_m *DB) StoreLastProcessedBlock(ctx context.Context, task string, block uint64, dbTx sqlx.ExecerContext) error {
	ret := _m.Called(ctx, task, block, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for StoreLastProcessedBlock")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint64, sqlx.ExecerContext) error); ok {
		r0 = rf(ctx, task, block, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreOffChainData provides a mock function with given fields: ctx, od, dbTx
func (_m *DB) StoreOffChainData(ctx context.Context, od []types.OffChainData, dbTx sqlx.ExecerContext) error {
	ret := _m.Called(ctx, od, dbTx)

	if len(ret) == 0 {
		panic("no return value specified for StoreOffChainData")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []types.OffChainData, sqlx.ExecerContext) error); ok {
		r0 = rf(ctx, od, dbTx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDB creates a new instance of DB. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDB(t interface {
	mock.TestingT
	Cleanup(func())
}) *DB {
	mock := &DB{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
