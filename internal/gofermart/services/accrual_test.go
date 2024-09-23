package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			name:            "Request_Failed_Invalid_Order",
			orderNumber:     "invalid_order_number",
			mockStatusCode:  http.StatusOK,
			mockResponse:    "",
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   fmt.Errorf("invalid order number: invalid_order_number"),
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
			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestProcessResponse - тестирует логику обработки ответа в GetOrderAccrual
func TestProcessResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()

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
			name:            "Invalid_JSON_Response",
			orderNumber:     "123456",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":"PROCESSED"}`, // Неполный ответ
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusProcessed,
			expectedError:   nil,
		},
		{
			name:            "Order_Processing",
			orderNumber:     "654321",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"654321","status":"PROCESSING","accrual":0}`,
			expectedAccrual: 0,
			expectedStatus:  domain.OrderStatusProcessing,
			expectedError:   nil,
		},
		{
			name:            "Order_Invalid",
			orderNumber:     "000000",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"000000","status":"INVALID"}`,
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

			// Инициализация сервиса с mock-адресом
			service := &RealAccrualService{
				BaseURL: mockServer.URL,
				logger:  logger,
			}

			// Вызываем метод для получения данных
			accrual, status, err := service.GetOrderAccrual(tc.orderNumber)

			// Проверяем результат
			if accrual != tc.expectedAccrual {
				t.Errorf("Expected accrual %v, got %v", tc.expectedAccrual, accrual)
			}
			if status != tc.expectedStatus {
				t.Errorf("Expected status %v, got %v", tc.expectedStatus, status)
			}
			if tc.expectedError != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
