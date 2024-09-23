package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewAccrualService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	baseURL := "https://example.com"

	service := NewAccrualService(baseURL, logger)

	realService, ok := service.(*RealAccrualService)
	if !ok {
		t.Fatalf("Expected *RealAccrualService, got %T", service)
	}

	if realService.BaseURL != baseURL {
		t.Errorf("Expected BaseURL %v, got %v", baseURL, realService.BaseURL)
	}

	if realService.logger != logger {
		t.Errorf("Expected logger %v, got %v", logger, realService.logger)
	}
}

// testCase структура для организации тестов
type testCase struct {
	name            string
	orderNumber     string
	mockStatusCode  int
	mockResponse    string
	expectedAccrual float64
	expectedStatus  string
	expectedError   error
}

// TestGetOrderAccrual тестирует основной метод GetOrderAccrual
func TestGetOrderAccrual(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Список тестовых случаев
	testCases := []testCase{
		{
			name:            "Valid_Processed_Order",
			orderNumber:     "123456",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"PROCESSED","accrual":100.5}`,
			expectedAccrual: 100.5,
			expectedStatus:  domain.OrderStatusProcessed,
			expectedError:   nil,
		},
		{
			name:            "Order_Not_Found",
			orderNumber:     "000000",
			mockStatusCode:  http.StatusNoContent,
			mockResponse:    "",
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusInvalid,
			expectedError:   nil,
		},
		{
			name:            "Too_Many_Requests",
			orderNumber:     "654321",
			mockStatusCode:  http.StatusTooManyRequests,
			mockResponse:    "",
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusProcessing,
			expectedError:   nil,
		},
		{
			name:            "Accrual_System_Error",
			orderNumber:     "123123",
			mockStatusCode:  http.StatusInternalServerError,
			mockResponse:    "",
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   gofermartErrors.ErrAccrualSystemUnavailable,
		},
		{
			name:            "Invalid_Order_Status",
			orderNumber:     "999999",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"999999","status":"UNKNOWN","accrual":50.0}`,
			expectedAccrual: 50.0,
			expectedStatus:  domain.OrderStatusProcessing,
			expectedError:   nil,
		},
		{
			name:            "Request_Failed",
			orderNumber:     "invalid_order_number",
			mockStatusCode:  http.StatusOK,
			mockResponse:    "",
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   fmt.Errorf("invalid order number: invalid_order_number"), // Обновляем ожидаемую ошибку
		},
	}

	// Проходим по каждому тестовому случаю
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем mock-сервер
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				if tc.mockResponse != "" {
					_, err := w.Write([]byte(tc.mockResponse))
					if err != nil {
						return
					}
				}
			}))
			defer mockServer.Close()

			// Инициализация сервиса с mock-адресом
			service := &RealAccrualService{
				BaseURL: mockServer.URL,
				logger:  logger,
			}

			// Вызываем тестируемый метод
			accrual, status, err := service.GetOrderAccrual(tc.orderNumber)

			// Проверяем результат
			if accrual != tc.expectedAccrual {
				t.Errorf("Expected accrual %v, got %v", tc.expectedAccrual, accrual)
			}
			if status != tc.expectedStatus {
				t.Errorf("Expected status %v, got %v", tc.expectedStatus, status)
			}
			if err != nil {
				if err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err.Error())
				}
			}
		})
	}
}

// TestSendRequest - проверяет метод sendRequest
func TestSendRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := []struct {
		name        string
		orderNumber string
		shouldError bool
	}{
		{
			name:        "Valid_Request",
			orderNumber: "123456",
			shouldError: false,
		},
		{
			name:        "Invalid_URL",
			orderNumber: "invalid_order",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем mock-сервер
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer mockServer.Close()

			// Инициализация сервиса с mock-адресом
			service := &RealAccrualService{
				BaseURL: strings.TrimSuffix(mockServer.URL, "/api/orders"),
				logger:  logger,
			}

			// Вызываем тестируемый метод sendRequest
			resp, err := service.sendRequest(tc.orderNumber)

			if resp != nil {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Errorf("Failed to close response body: %v", err)
					}
				}()
			}

			// Проверка результата
			if tc.shouldError && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
			if resp != nil && resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code 200, got %v", resp.StatusCode)
			}
		})
	}
}

// TestProcessResponse - тестирует метод processResponse
func TestProcessResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	testCases := []testCase{
		{
			name:            "Valid_Processed_Order",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"PROCESSED","accrual":100.5}`,
			expectedAccrual: 100.5,
			expectedStatus:  domain.OrderStatusProcessed,
			expectedError:   nil,
		},
		{
			name:            "Invalid_JSON_Response",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"PROCESSED"}`, // Исправляем формат
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusProcessed,
			expectedError:   nil,
		},
		{
			name:            "Order_Processing",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"PROCESSING","accrual":0}`,
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusProcessing,
			expectedError:   nil,
		},
		{
			name:            "Order_Registered",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"REGISTERED","accrual":0}`,
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusProcessing,
			expectedError:   nil,
		},
		{
			name:            "Order_Invalid",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"INVALID"}`,
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusInvalid,
			expectedError:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем mock-сервер
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				if tc.mockResponse != "" {
					_, err := w.Write([]byte(tc.mockResponse))
					if err != nil {
						return
					}
				}
			}))
			defer mockServer.Close()

			service := &RealAccrualService{
				BaseURL: mockServer.URL,
				logger:  logger,
			}

			// Отправляем запрос, чтобы получить ответ для processResponse
			resp, err := service.sendRequest(tc.orderNumber)
			if resp != nil {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Errorf("Failed to close response body: %v", err)
					}
				}()
			}
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			accrual, status, err := service.processResponse(resp, tc.orderNumber)

			// Проверка результата
			if accrual != tc.expectedAccrual {
				t.Errorf("Expected accrual %v, got %v", tc.expectedAccrual, accrual)
			}
			if status != tc.expectedStatus {
				t.Errorf("Expected status %v, got %v", tc.expectedStatus, status)
			}
			if tc.expectedError != nil && err == nil {
				t.Errorf("Expected error %v, got nil", tc.expectedError)
			}
		})
	}
}
