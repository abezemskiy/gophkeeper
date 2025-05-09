// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/abezemskiy/gophkeeper/internal/client/storage (interfaces: IEncryptedClientStorage)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	data "github.com/abezemskiy/gophkeeper/internal/repositories/data"
	gomock "github.com/golang/mock/gomock"
)

// MockIEncryptedClientStorage is a mock of IEncryptedClientStorage interface.
type MockIEncryptedClientStorage struct {
	ctrl     *gomock.Controller
	recorder *MockIEncryptedClientStorageMockRecorder
}

// MockIEncryptedClientStorageMockRecorder is the mock recorder for MockIEncryptedClientStorage.
type MockIEncryptedClientStorageMockRecorder struct {
	mock *MockIEncryptedClientStorage
}

// NewMockIEncryptedClientStorage creates a new mock instance.
func NewMockIEncryptedClientStorage(ctrl *gomock.Controller) *MockIEncryptedClientStorage {
	mock := &MockIEncryptedClientStorage{ctrl: ctrl}
	mock.recorder = &MockIEncryptedClientStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIEncryptedClientStorage) EXPECT() *MockIEncryptedClientStorageMockRecorder {
	return m.recorder
}

// AddEncryptedData mocks base method.
func (m *MockIEncryptedClientStorage) AddEncryptedData(arg0 context.Context, arg1 string, arg2 data.EncryptedData, arg3 int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddEncryptedData", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddEncryptedData indicates an expected call of AddEncryptedData.
func (mr *MockIEncryptedClientStorageMockRecorder) AddEncryptedData(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddEncryptedData", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).AddEncryptedData), arg0, arg1, arg2, arg3)
}

// ChangeStatusOfEncryptedData mocks base method.
func (m *MockIEncryptedClientStorage) ChangeStatusOfEncryptedData(arg0 context.Context, arg1, arg2 string, arg3 int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChangeStatusOfEncryptedData", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChangeStatusOfEncryptedData indicates an expected call of ChangeStatusOfEncryptedData.
func (mr *MockIEncryptedClientStorageMockRecorder) ChangeStatusOfEncryptedData(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChangeStatusOfEncryptedData", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).ChangeStatusOfEncryptedData), arg0, arg1, arg2, arg3)
}

// DeleteEncryptedData mocks base method.
func (m *MockIEncryptedClientStorage) DeleteEncryptedData(arg0 context.Context, arg1, arg2 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEncryptedData", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteEncryptedData indicates an expected call of DeleteEncryptedData.
func (mr *MockIEncryptedClientStorageMockRecorder) DeleteEncryptedData(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEncryptedData", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).DeleteEncryptedData), arg0, arg1, arg2)
}

// GetAllEncryptedData mocks base method.
func (m *MockIEncryptedClientStorage) GetAllEncryptedData(arg0 context.Context, arg1 string) ([][]data.EncryptedData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllEncryptedData", arg0, arg1)
	ret0, _ := ret[0].([][]data.EncryptedData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllEncryptedData indicates an expected call of GetAllEncryptedData.
func (mr *MockIEncryptedClientStorageMockRecorder) GetAllEncryptedData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllEncryptedData", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).GetAllEncryptedData), arg0, arg1)
}

// GetEncryptedDataByStatus mocks base method.
func (m *MockIEncryptedClientStorage) GetEncryptedDataByStatus(arg0 context.Context, arg1 string, arg2 int) ([][]data.EncryptedData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEncryptedDataByStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].([][]data.EncryptedData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEncryptedDataByStatus indicates an expected call of GetEncryptedDataByStatus.
func (mr *MockIEncryptedClientStorageMockRecorder) GetEncryptedDataByStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEncryptedDataByStatus", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).GetEncryptedDataByStatus), arg0, arg1, arg2)
}

// GetStatus mocks base method.
func (m *MockIEncryptedClientStorage) GetStatus(arg0 context.Context, arg1, arg2 string) (int, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetStatus indicates an expected call of GetStatus.
func (mr *MockIEncryptedClientStorageMockRecorder) GetStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatus", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).GetStatus), arg0, arg1, arg2)
}

// ReplaceDataWithMultiVersionData mocks base method.
func (m *MockIEncryptedClientStorage) ReplaceDataWithMultiVersionData(arg0 context.Context, arg1 string, arg2 []data.EncryptedData, arg3 int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReplaceDataWithMultiVersionData", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReplaceDataWithMultiVersionData indicates an expected call of ReplaceDataWithMultiVersionData.
func (mr *MockIEncryptedClientStorageMockRecorder) ReplaceDataWithMultiVersionData(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReplaceDataWithMultiVersionData", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).ReplaceDataWithMultiVersionData), arg0, arg1, arg2, arg3)
}

// ReplaceEncryptedData mocks base method.
func (m *MockIEncryptedClientStorage) ReplaceEncryptedData(arg0 context.Context, arg1 string, arg2 data.EncryptedData, arg3 int) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReplaceEncryptedData", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReplaceEncryptedData indicates an expected call of ReplaceEncryptedData.
func (mr *MockIEncryptedClientStorageMockRecorder) ReplaceEncryptedData(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReplaceEncryptedData", reflect.TypeOf((*MockIEncryptedClientStorage)(nil).ReplaceEncryptedData), arg0, arg1, arg2, arg3)
}
