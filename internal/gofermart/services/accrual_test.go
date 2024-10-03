package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"context"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

	if realService.limiter == nil {
		t.Error("Expected limiter to be initialized")
	} else if realService.limiter.Limit() != rate.Inf {
		t.Errorf("Expected limiter limit to be rate.Inf, got %v", realService.limiter.Limit())
	}
}

func TestGetOrderAccrual(t *testing.T) {
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
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   errors.New("received unknown order status from the accrual system"),
		},
		{
			name:            "Failed_to_Create_New_Request",
			orderNumber:     "\x7f",
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   errors.New("failed to create new request"),
		},
		{
			name:            "Failed_to_Decode_JSON_Response",
			orderNumber:     "invalid_json",
			mockStatusCode:  http.StatusOK,
			mockResponse:    `{"order":"123456","status":123}`,
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   errors.New("failed to decode JSON response"),
		},
		{
			name:            "Invalid_URL_Request",
			orderNumber:     "request_failure",
			expectedAccrual: 0,
			expectedStatus:  "",
			expectedError:   errors.New("unsupported protocol scheme"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем mock-сервер
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.mockStatusCode == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", "0")
				}
				w.WriteHeader(tc.mockStatusCode)
				if tc.mockResponse != "" {
					_, err := w.Write([]byte(tc.mockResponse))
					if err != nil {
						t.Fatalf("Failed to write mock response: %v", err)
					}
				}
			}))
			defer mockServer.Close()

			accrualURL := mockServer.URL
			if tc.name == "Invalid_URL_Request" {
				accrualURL = "invalid-url"
			}

			service := &RealAccrualService{
				BaseURL: accrualURL,
				logger:  logger,
				limiter: rate.NewLimiter(rate.Inf, 1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			accrual, status, err := service.GetOrderAccrual(ctx, tc.orderNumber)

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
					t.Errorf("Expected error containing %v, got %v", tc.expectedError.Error(), err.Error())
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
