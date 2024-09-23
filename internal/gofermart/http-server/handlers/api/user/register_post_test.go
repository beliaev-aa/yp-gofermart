package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"bytes"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterPostHandler_ServeHTTP(t *testing.T) {
	logger := zap.NewNop()

	mockStorage := &tests.MockStorage{
		SaveUserFn: func(user domain.User) error {
			if user.Login == "existing_user" {
				return gofermartErrors.ErrLoginAlreadyExists
			} else if user.Login == "server_error" {
				return errors.New("user not found")
			}
			return nil
		},
		GetUserByLoginFn: func(login string) (*domain.User, error) {
			if login == "existing_user" {
				return &domain.User{Login: login}, nil
			}
			return nil, nil
		},
	}

	authService := services.NewAuthService([]byte("secret"), logger, mockStorage)

	handler := NewRegisterPostHandler(authService, logger)

	testCases := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{
			name: "Success_Response",
			requestBody: `{
				"login": "new_user",
				"password": "password123"
			}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Login_Already_Exists",
			requestBody: `{
				"login": "existing_user",
				"password": "password123"
			}`,
			expectedStatusCode: http.StatusConflict,
		},
		{
			name: "Server_Error",
			requestBody: `{
				"login": "server_error",
				"password": "password123"
			}`,
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
			req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(tc.requestBody)))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if tc.name == "Invalid_Request_Format" && rr.Body.String() != "Invalid request format\n" {
				t.Errorf("expected error message 'Invalid request format', got %v", rr.Body.String())
			}
		})
	}
}
