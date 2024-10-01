package tests

import (
	"go.uber.org/zap"
	"net/http"
)

// MockUsernameExtractor - mock реализация интерфейса UsernameExtractor
type MockUsernameExtractor struct {
	ExtractFn func(r *http.Request, logger *zap.Logger) (string, error)
}

func (m *MockUsernameExtractor) ExtractUsernameFromContext(r *http.Request, logger *zap.Logger) (string, error) {
	return m.ExtractFn(r, logger)
}
