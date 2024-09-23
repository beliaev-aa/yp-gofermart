package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRealAccrualService_GetOrderAccrual(t *testing.T) {
	testCases := []struct {
		Name            string
		OrderNumber     string
		ServerResponse  interface{}
		ServerStatus    int
		ExpectedAccrual float64
		ExpectedStatus  string
		ExpectedError   string
	}{
		{
			Name:            "GetOrderAccrual_Success_Processed",
			OrderNumber:     "12345",
			ServerResponse:  map[string]interface{}{"order": "12345", "status": "PROCESSED", "accrual": 100.0},
			ServerStatus:    http.StatusOK,
			ExpectedAccrual: 100.0,
			ExpectedStatus:  domain.OrderStatusProcessed,
			ExpectedError:   "",
		},
		{
			Name:            "GetOrderAccrual_Success_Processing",
			OrderNumber:     "12345",
			ServerResponse:  map[string]interface{}{"order": "12345", "status": "PROCESSING"},
			ServerStatus:    http.StatusOK,
			ExpectedAccrual: 0.0,
			ExpectedStatus:  domain.OrderStatusProcessing,
			ExpectedError:   "",
		},
		{
			Name:            "GetOrderAccrual_Success_Invalid",
			OrderNumber:     "12345",
			ServerResponse:  map[string]interface{}{"order": "12345", "status": "INVALID"},
			ServerStatus:    http.StatusOK,
			ExpectedAccrual: 0.0,
			ExpectedStatus:  domain.OrderStatusInvalid,
			ExpectedError:   "",
		},
		{
			Name:            "GetOrderAccrual_TooManyRequests",
			OrderNumber:     "12345",
			ServerResponse:  nil,
			ServerStatus:    http.StatusTooManyRequests,
			ExpectedAccrual: 0.0,
			ExpectedStatus:  domain.OrderStatusProcessing,
			ExpectedError:   "",
		},
		{
			Name:            "GetOrderAccrual_NoContent",
			OrderNumber:     "12345",
			ServerResponse:  nil,
			ServerStatus:    http.StatusNoContent,
			ExpectedAccrual: 0.0,
			ExpectedStatus:  domain.OrderStatusInvalid,
			ExpectedError:   "",
		},
		{
			Name:            "GetOrderAccrual_AccrualSystemUnavailable",
			OrderNumber:     "12345",
			ServerResponse:  nil,
			ServerStatus:    http.StatusInternalServerError,
			ExpectedAccrual: 0.0,
			ExpectedStatus:  "",
			ExpectedError:   gofermartErrors.ErrAccrualSystemUnavailable.Error(),
		},
		{
			Name:            "GetOrderAccrual_InvalidJSON",
			OrderNumber:     "12345",
			ServerResponse:  `{"order": "12345", "status": "PROCESSED", "accrual": "invalid-accrual"}`, // Неверный формат для поля Accrual
			ServerStatus:    http.StatusOK,
			ExpectedAccrual: 0.0,
			ExpectedStatus:  "",
			ExpectedError:   "cannot unmarshal string into Go struct field", // Используем ключевые слова для проверки
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Создаем тестовый сервер
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.ServerStatus)
				if tc.ServerResponse != nil {
					switch v := tc.ServerResponse.(type) {
					case string:
						_, err := w.Write([]byte(v))
						if err != nil {
							return
						}
					default:
						err := json.NewEncoder(w).Encode(tc.ServerResponse)
						if err != nil {
							return
						}
					}
				}
			}))
			defer ts.Close()

			// Создаем новый экземпляр RealAccrualService
			logger := zap.NewNop()
			accrualService := NewAccrualService(ts.URL, logger)

			// Вызываем метод GetOrderAccrual
			accrual, status, err := accrualService.GetOrderAccrual(tc.OrderNumber)

			// Проверяем результат
			if accrual != tc.ExpectedAccrual {
				t.Errorf("Expected accrual %v, got %v", tc.ExpectedAccrual, accrual)
			}
			if status != tc.ExpectedStatus {
				t.Errorf("Expected status %v, got %v", tc.ExpectedStatus, status)
			}

			if tc.ExpectedError == "" && err != nil {
				t.Errorf("Expected no error, got %v", err)
			} else if tc.ExpectedError != "" && err == nil {
				t.Errorf("Expected error %v, got none", tc.ExpectedError)
			} else if err != nil && !strings.Contains(err.Error(), tc.ExpectedError) {
				t.Errorf("Expected error to contain %v, got %v", tc.ExpectedError, err.Error())
			}
		})
	}
}
