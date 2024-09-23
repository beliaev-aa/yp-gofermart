package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"bytes"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoginPostHandler_ServeHTTP(t *testing.T) {
	logger := zap.NewNop()

	testCases := []struct {
		name               string
		requestBody        string
		mockGetUserByLogin func(login string) (*domain.User, error)
		mockGenerateJWT    func(login string) (string, error)
		expectedStatusCode int
	}{
		{
			name: "Authentication_Failed",
			requestBody: `{
				"login": "test_user",
				"password": "wrong_password"
			}`,
			mockGetUserByLogin: func(login string) (*domain.User, error) {
				return &domain.User{Login: "test_user", Password: "password123"}, nil
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Server_Error_During_Authentication",
			requestBody: `{
				"login": "test_user",
				"password": "password123"
			}`,
			mockGetUserByLogin: func(login string) (*domain.User, error) {
				return nil, errors.New("internal error")
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:               "Invalid_Request_Format",
			requestBody:        `{invalid json}`,
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage := &tests.MockStorage{
				GetUserByLoginFn: tc.mockGetUserByLogin,
			}

			authService := services.NewAuthService([]byte("secret"), logger, mockStorage)

			handler := NewLoginPostHandler(authService, logger)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(tc.requestBody)))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if tc.expectedStatusCode == http.StatusOK {
				authHeader := rr.Header().Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					t.Errorf("expected Authorization header to start with 'Bearer ', got %v", authHeader)
				}
			}
		})
	}
}
