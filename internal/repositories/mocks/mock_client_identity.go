// Code generated by MockGen. DO NOT EDIT.
// Source: gophkeeper/internal/client/identity (interfaces: ClientIdentifier)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	identity "gophkeeper/internal/client/identity"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockClientIdentifier is a mock of ClientIdentifier interface.
type MockClientIdentifier struct {
	ctrl     *gomock.Controller
	recorder *MockClientIdentifierMockRecorder
}

// MockClientIdentifierMockRecorder is the mock recorder for MockClientIdentifier.
type MockClientIdentifierMockRecorder struct {
	mock *MockClientIdentifier
}

// NewMockClientIdentifier creates a new mock instance.
func NewMockClientIdentifier(ctrl *gomock.Controller) *MockClientIdentifier {
	mock := &MockClientIdentifier{ctrl: ctrl}
	mock.recorder = &MockClientIdentifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClientIdentifier) EXPECT() *MockClientIdentifierMockRecorder {
	return m.recorder
}

// Authorize mocks base method.
func (m *MockClientIdentifier) Authorize(arg0 context.Context, arg1, arg2 string) (identity.UserInfo, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", arg0, arg1, arg2)
	ret0, _ := ret[0].(identity.UserInfo)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Authorize indicates an expected call of Authorize.
func (mr *MockClientIdentifierMockRecorder) Authorize(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockClientIdentifier)(nil).Authorize), arg0, arg1, arg2)
}

// Register mocks base method.
func (m *MockClientIdentifier) Register(arg0 context.Context, arg1, arg2, arg3, arg4 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockClientIdentifierMockRecorder) Register(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockClientIdentifier)(nil).Register), arg0, arg1, arg2, arg3, arg4)
}

// SetToken mocks base method.
func (m *MockClientIdentifier) SetToken(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetToken", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetToken indicates an expected call of SetToken.
func (mr *MockClientIdentifierMockRecorder) SetToken(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetToken", reflect.TypeOf((*MockClientIdentifier)(nil).SetToken), arg0, arg1, arg2)
}
