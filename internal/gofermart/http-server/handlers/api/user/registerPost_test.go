package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
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

func TestRegisterPostHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	logger := zap.NewNop()
	jwtSecret := []byte("secret")
	authService := services.NewAuthService(jwtSecret, mockUserRepo, logger)
	handler := NewRegisterPostHandler(authService, logger)

	testCases := []struct {
		Name                 string
		RequestBody          string
		SetupMocks           func()
		ExpectedStatusCode   int
		ExpectedResponseBody string
		ExpectedAuthHeader   string
	}{
		{
			Name:                 "Invalid_Request_Format",
			RequestBody:          `{invalid json}`,
			SetupMocks:           func() {},
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedResponseBody: "Invalid request format\n",
		},
		{
			Name:        "Registration_Conflict",
			RequestBody: `{"login": "user1", "password": "password123"}`,
			SetupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{Login: "user1"}, nil)
			},
			ExpectedStatusCode:   http.StatusConflict,
			ExpectedResponseBody: "login already exist\n",
		},
		{
			Name:        "Registration_Server_Error",
			RequestBody: `{"login": "user1", "password": "password123"}`,
			SetupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(nil, gofermartErrors.ErrUserNotFound)
				mockUserRepo.EXPECT().SaveUser(gomock.Any()).Return(errors.New("db error")) // Добавлено ожидание вызова SaveUser
			},
			ExpectedStatusCode:   http.StatusInternalServerError,
			ExpectedResponseBody: "Server error\n",
		},
		{
			Name:        "Successful_Registration",
			RequestBody: `{"login": "user1", "password": "password123"}`,
			SetupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(nil, gofermartErrors.ErrUserNotFound)
				mockUserRepo.EXPECT().SaveUser(gomock.Any()).Return(nil)
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedAuthHeader: "Bearer ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.SetupMocks()

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tc.RequestBody))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.ExpectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tc.ExpectedStatusCode, rec.Code)
			}

			if rec.Body.String() != tc.ExpectedResponseBody {
				t.Errorf("Expected response body '%s', got '%s'", tc.ExpectedResponseBody, rec.Body.String())
			}

			if tc.ExpectedStatusCode == http.StatusOK {
				authHeader := rec.Header().Get("Authorization")
				if !bytes.HasPrefix([]byte(authHeader), []byte(tc.ExpectedAuthHeader)) {
					t.Errorf("Expected Authorization header to start with '%s', got '%s'", tc.ExpectedAuthHeader, authHeader)
				}
			}
		})
	}
}
