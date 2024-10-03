package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"bytes"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginPostHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()
	jwtSecret := []byte("secret")
	authService := services.NewAuthService(jwtSecret, mockUserRepo, logger)
	handler := NewLoginPostHandler(authService, logger)

	testCases := []struct {
		name                 string
		requestBody          string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
		expectedAuthHeader   string
	}{
		{
			name:                 "Invalid_Request_Format",
			requestBody:          `{invalid json}`,
			setupMocks:           func() {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "Invalid request format\n",
		},
		{
			name:        "Authentication_Error",
			requestBody: `{"login": "user1", "password": "wrong_password"}`,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(nil, errors.New("db error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Server error\n",
		},
		{
			name:        "Invalid_Login_Or_Password",
			requestBody: `{"login": "user1", "password": "wrong_password"}`,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{Login: "user1", Password: "hashed_password"}, nil)
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: "Invalid login/password\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tc.requestBody))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			if rec.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatusCode, rec.Code)
			}

			if rec.Body.String() != tc.expectedResponseBody {
				t.Errorf("Expected response body '%s', got '%s'", tc.expectedResponseBody, rec.Body.String())
			}

			if tc.expectedStatusCode == http.StatusOK {
				authHeader := rec.Header().Get("Authorization")
				if !bytes.HasPrefix([]byte(authHeader), []byte(tc.expectedAuthHeader)) {
					t.Errorf("Expected Authorization header to start with '%s', got '%s'", tc.expectedAuthHeader, authHeader)
				}
			}
		})
	}
}
