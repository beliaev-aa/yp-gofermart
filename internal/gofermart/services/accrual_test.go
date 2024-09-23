package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestNewAccrualService(t *testing.T) {
	logger := zap.NewNop()
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

// TestGetOrderAccrual тестирует основной метод GetOrderAccrual
func TestGetOrderAccrual(t *testing.T) {
	logger := zap.NewNop()

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
			name:            "Failed_to_Create_New_Request",
			orderNumber:     "\x7f", // Некорректный символ в номере заказа
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   errors.New("invalid control character in URL"),
		},
		{
			name:            "Failed_to_Decode_JSON_Response",
			orderNumber:     "invalid_json",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":123}`, // Некорректный тип данных для статуса
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   errors.New("json: cannot unmarshal number into Go struct field"),
		},
	}

	// Проходим по каждому тестовому случаю
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем mock-сервер
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.orderNumber == "request_error" {
					// Симулируем сбой выполнения запроса
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
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
				} else if !strings.Contains(err.Error(), tc.expectedError.Error()) {
					t.Errorf("Expected error %v, got %v", tc.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

// TestProcessResponse - тестирует логику обработки ответа в GetOrderAccrual
func TestProcessResponse(t *testing.T) {
	logger := zap.NewNop()

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
