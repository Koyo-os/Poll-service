// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Casher is an autogenerated mock type for the Casher type
type Casher struct {
	mock.Mock
}

// DoCashing provides a mock function with given fields: ctx, key, payload
func (_m *Casher) DoCashing(ctx context.Context, key string, payload interface{}) error {
	ret := _m.Called(ctx, key, payload)

	if len(ret) == 0 {
		panic("no return value specified for DoCashing")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) error); ok {
		r0 = rf(ctx, key, payload)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewCasher creates a new instance of Casher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCasher(t interface {
	mock.TestingT
	Cleanup(func())
}) *Casher {
	mock := &Casher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
