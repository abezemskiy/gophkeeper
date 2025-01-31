// Code generated by MockGen. DO NOT EDIT.
// Source: gophkeeper/internal/server/storage (interfaces: IEncryptedServerStorage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	data "gophkeeper/internal/repositories/data"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockIEncryptedServerStorage is a mock of IEncryptedServerStorage interface.
type MockIEncryptedServerStorage struct {
	ctrl     *gomock.Controller
	recorder *MockIEncryptedServerStorageMockRecorder
}

// MockIEncryptedServerStorageMockRecorder is the mock recorder for MockIEncryptedServerStorage.
type MockIEncryptedServerStorageMockRecorder struct {
	mock *MockIEncryptedServerStorage
}

// NewMockIEncryptedServerStorage creates a new mock instance.
func NewMockIEncryptedServerStorage(ctrl *gomock.Controller) *MockIEncryptedServerStorage {
	mock := &MockIEncryptedServerStorage{ctrl: ctrl}
	mock.recorder = &MockIEncryptedServerStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIEncryptedServerStorage) EXPECT() *MockIEncryptedServerStorageMockRecorder {
	return m.recorder
}

// AddDateOfLastVisit mocks base method.
func (m *MockIEncryptedServerStorage) AddDateOfLastVisit(arg0 context.Context, arg1 string, arg2 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddDateOfLastVisit", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddDateOfLastVisit indicates an expected call of AddDateOfLastVisit.
func (mr *MockIEncryptedServerStorageMockRecorder) AddDateOfLastVisit(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddDateOfLastVisit", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).AddDateOfLastVisit), arg0, arg1, arg2)
}

// AddEncryptedData mocks base method.
func (m *MockIEncryptedServerStorage) AddEncryptedData(arg0 context.Context, arg1 string, arg2 data.EncryptedData) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddEncryptedData", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddEncryptedData indicates an expected call of AddEncryptedData.
func (mr *MockIEncryptedServerStorageMockRecorder) AddEncryptedData(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEncryptedData", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).AddEncryptedData), arg0, arg1, arg2)
}

// AppendEncryptedData mocks base method.
func (m *MockIEncryptedServerStorage) AppendEncryptedData(arg0 context.Context, arg1 string, arg2 data.EncryptedData) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AppendEncryptedData", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AppendEncryptedData indicates an expected call of AppendEncryptedData.
func (mr *MockIEncryptedServerStorageMockRecorder) AppendEncryptedData(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendEncryptedData", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).AppendEncryptedData), arg0, arg1, arg2)
}

// Bootstrap mocks base method.
func (m *MockIEncryptedServerStorage) Bootstrap(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bootstrap", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Bootstrap indicates an expected call of Bootstrap.
func (mr *MockIEncryptedServerStorageMockRecorder) Bootstrap(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bootstrap", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).Bootstrap), arg0)
}

// GetAllEncryptedData mocks base method.
func (m *MockIEncryptedServerStorage) GetAllEncryptedData(arg0 context.Context, arg1 string) ([][]data.EncryptedData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllEncryptedData", arg0, arg1)
	ret0, _ := ret[0].([][]data.EncryptedData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllEncryptedData indicates an expected call of GetAllEncryptedData.
func (mr *MockIEncryptedServerStorageMockRecorder) GetAllEncryptedData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllEncryptedData", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).GetAllEncryptedData), arg0, arg1)
}

// GetDateOfLastVisit mocks base method.
func (m *MockIEncryptedServerStorage) GetDateOfLastVisit(arg0 context.Context, arg1 string) (time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDateOfLastVisit", arg0, arg1)
	ret0, _ := ret[0].(time.Time)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDateOfLastVisit indicates an expected call of GetDateOfLastVisit.
func (mr *MockIEncryptedServerStorageMockRecorder) GetDateOfLastVisit(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDateOfLastVisit", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).GetDateOfLastVisit), arg0, arg1)
}

// ReplaceEncryptedData mocks base method.
func (m *MockIEncryptedServerStorage) ReplaceEncryptedData(arg0 context.Context, arg1 string, arg2 data.EncryptedData) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReplaceEncryptedData", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReplaceEncryptedData indicates an expected call of ReplaceEncryptedData.
func (mr *MockIEncryptedServerStorageMockRecorder) ReplaceEncryptedData(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReplaceEncryptedData", reflect.TypeOf((*MockIEncryptedServerStorage)(nil).ReplaceEncryptedData), arg0, arg1, arg2)
}
