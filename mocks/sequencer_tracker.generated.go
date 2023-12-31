// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	sequencer "github.com/0xPolygon/cdk-data-availability/sequencer"
	mock "github.com/stretchr/testify/mock"
)

// ISequencerTracker is an autogenerated mock type for the ISequencerTracker type
type ISequencerTracker struct {
	mock.Mock
}

// GetSequenceBatch provides a mock function with given fields: batchNum
func (_m *ISequencerTracker) GetSequenceBatch(batchNum uint64) (*sequencer.SeqBatch, error) {
	ret := _m.Called(batchNum)

	if len(ret) == 0 {
		panic("no return value specified for GetSequenceBatch")
	}

	var r0 *sequencer.SeqBatch
	var r1 error
	if rf, ok := ret.Get(0).(func(uint64) (*sequencer.SeqBatch, error)); ok {
		return rf(batchNum)
	}
	if rf, ok := ret.Get(0).(func(uint64) *sequencer.SeqBatch); ok {
		r0 = rf(batchNum)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*sequencer.SeqBatch)
		}
	}

	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(batchNum)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewISequencerTracker creates a new instance of ISequencerTracker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewISequencerTracker(t interface {
	mock.TestingT
	Cleanup(func())
}) *ISequencerTracker {
	mock := &ISequencerTracker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
