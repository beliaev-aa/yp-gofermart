package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestRegisterPostHandler_ServeHTTP(t *testing.T) {
	logger := zap.NewNop()

	// Используем MockStorage
	mockStorage := &tests.MockStorage{
		SaveUserFn: func(user domain.User) error {
			if user.Login == "existing_user" {
				return errors.ErrLoginAlreadyExists
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

	// Создаем AuthService через конструктор
	authService := services.NewAuthService([]byte("secret"), logger, mockStorage)

	// Создаем тестируемый обработчик
	handler := NewRegisterPostHandler(authService, logger)

	// Тестовые данные
	testCases := []struct {
		name               string
		requestBody        domain.AuthenticationRequest
		expectedStatusCode int
	}{
		{
			name: "Success_Response",
			requestBody: domain.AuthenticationRequest{
				Login:    "new_user",
				Password: "password123",
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "Login_Already_Exists",
			requestBody: domain.AuthenticationRequest{
				Login:    "existing_user",
				Password: "password123",
			},
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем тело запроса
			requestBody, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/register", bytes.NewReader(requestBody))

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Вызываем ServeHTTP
			handler.ServeHTTP(rr, req)

			// Проверяем статус ответа
			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}
		})
	}
}
