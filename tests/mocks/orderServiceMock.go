// Code generated by MockGen. DO NOT EDIT.
// Source: internal/gofermart/services/order.go

// Package mocks is a generated GoMock package.
package mocks

import (
	domain "beliaev-aa/yp-gofermart/internal/gofermart/domain"
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	decimal "github.com/shopspring/decimal"
	gorm "gorm.io/gorm"
)

// MockOrderServiceInterface is a mock of OrderServiceInterface interface.
type MockOrderServiceInterface struct {
	ctrl     *gomock.Controller
	recorder *MockOrderServiceInterfaceMockRecorder
}

// MockOrderServiceInterfaceMockRecorder is the mock recorder for MockOrderServiceInterface.
type MockOrderServiceInterfaceMockRecorder struct {
	mock *MockOrderServiceInterface
}

// NewMockOrderServiceInterface creates a new mock instance.
func NewMockOrderServiceInterface(ctrl *gomock.Controller) *MockOrderServiceInterface {
	mock := &MockOrderServiceInterface{ctrl: ctrl}
	mock.recorder = &MockOrderServiceInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOrderServiceInterface) EXPECT() *MockOrderServiceInterfaceMockRecorder {
	return m.recorder
}

// AddOrder mocks base method.
func (m *MockOrderServiceInterface) AddOrder(login, number string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddOrder", login, number)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddOrder indicates an expected call of AddOrder.
func (mr *MockOrderServiceInterfaceMockRecorder) AddOrder(login, number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddOrder", reflect.TypeOf((*MockOrderServiceInterface)(nil).AddOrder), login, number)
}

// GetOrders mocks base method.
func (m *MockOrderServiceInterface) GetOrders(login string) ([]domain.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", login)
	ret0, _ := ret[0].([]domain.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockOrderServiceInterfaceMockRecorder) GetOrders(login interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockOrderServiceInterface)(nil).GetOrders), login)
}

// UpdateOrderStatuses mocks base method.
func (m *MockOrderServiceInterface) UpdateOrderStatuses(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateOrderStatuses", ctx)
}

// UpdateOrderStatuses indicates an expected call of UpdateOrderStatuses.
func (mr *MockOrderServiceInterfaceMockRecorder) UpdateOrderStatuses(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrderStatuses", reflect.TypeOf((*MockOrderServiceInterface)(nil).UpdateOrderStatuses), ctx)
}

// UpdateUserBalance mocks base method.
func (m *MockOrderServiceInterface) UpdateUserBalance(tx *gorm.DB, userID int, amount decimal.Decimal) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUserBalance", tx, userID, amount)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUserBalance indicates an expected call of UpdateUserBalance.
func (mr *MockOrderServiceInterfaceMockRecorder) UpdateUserBalance(tx, userID, amount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUserBalance", reflect.TypeOf((*MockOrderServiceInterface)(nil).UpdateUserBalance), tx, userID, amount)
}
