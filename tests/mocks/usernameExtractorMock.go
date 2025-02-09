// Code generated by MockGen. DO NOT EDIT.
// Source: internal/gofermart/utils/jwt.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	zap "go.uber.org/zap"
	http "net/http"
	reflect "reflect"
)

// MockUsernameExtractor is a mock of UsernameExtractor interface.
type MockUsernameExtractor struct {
	ctrl     *gomock.Controller
	recorder *MockUsernameExtractorMockRecorder
}

// MockUsernameExtractorMockRecorder is the mock recorder for MockUsernameExtractor.
type MockUsernameExtractorMockRecorder struct {
	mock *MockUsernameExtractor
}

// NewMockUsernameExtractor creates a new mock instance.
func NewMockUsernameExtractor(ctrl *gomock.Controller) *MockUsernameExtractor {
	mock := &MockUsernameExtractor{ctrl: ctrl}
	mock.recorder = &MockUsernameExtractorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsernameExtractor) EXPECT() *MockUsernameExtractorMockRecorder {
	return m.recorder
}

// ExtractUsernameFromContext mocks base method.
func (m *MockUsernameExtractor) ExtractUsernameFromContext(r *http.Request, logger *zap.Logger) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExtractUsernameFromContext", r, logger)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExtractUsernameFromContext indicates an expected call of ExtractUsernameFromContext.
func (mr *MockUsernameExtractorMockRecorder) ExtractUsernameFromContext(r, logger interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExtractUsernameFromContext", reflect.TypeOf((*MockUsernameExtractor)(nil).ExtractUsernameFromContext), r, logger)
}
