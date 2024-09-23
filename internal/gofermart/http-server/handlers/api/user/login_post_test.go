package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestLoginPostHandler_ServeHTTP(t *testing.T) {
	logger := zap.NewNop()

	// Тестовые кейсы
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

			// Создаем AuthService через конструктор
			authService := services.NewAuthService([]byte("secret"), logger, mockStorage)

			// Создаем тестируемый обработчик
			handler := NewLoginPostHandler(authService, logger)

			// Создаем тело запроса
			req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(tc.requestBody)))

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Вызываем ServeHTTP
			handler.ServeHTTP(rr, req)

			// Проверяем статус ответа
			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			// Проверяем, что заголовок авторизации начинается с "Bearer ", если статус OK
			if tc.expectedStatusCode == http.StatusOK {
				authHeader := rr.Header().Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					t.Errorf("expected Authorization header to start with 'Bearer ', got %v", authHeader)
				}
			}
		})
	}
}
