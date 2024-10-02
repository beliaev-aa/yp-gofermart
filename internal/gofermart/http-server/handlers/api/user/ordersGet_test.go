package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOrdersGetHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := zap.NewNop()

	testCases := []struct {
		name               string
		mockExtractFn      func(r *http.Request, logger *zap.Logger) (string, error)
		orders             []domain.Order
		mockError          error
		expectedStatusCode int
		expectedResponse   []OrderResponse
	}{
		{
			name: "Unauthorized_Access",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "", http.ErrNoCookie
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Successful_Order_Response",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			orders: []domain.Order{
				{
					OrderNumber: "123",
					OrderStatus: domain.OrderStatusProcessed,
					Accrual:     150.5,
					UploadedAt:  time.Now(),
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: []OrderResponse{
				{
					Number:     "123",
					Status:     domain.OrderStatusProcessed,
					Accrual:    150.5,
					UploadedAt: time.Now().Format(time.RFC3339),
				},
			},
		},
		{
			name: "No_Orders",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			orders:             []domain.Order{},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "Internal_Server_Error",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockError:          errors.New("database error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	accrualMock := mocks.NewMockAccrualService(ctrl)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExtractor := &tests.MockUsernameExtractor{
				ExtractFn: tc.mockExtractFn,
			}

			mockOrderService := services.NewOrderService(accrualMock, &tests.MockStorage{
				GetUserByLoginFn: func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				},
				GetOrdersByUserIDFn: func(userID int) ([]domain.Order, error) {
					return tc.orders, tc.mockError
				},
			}, logger)

			handler := NewOrdersGetHandler(mockOrderService, mockExtractor, logger)

			req := httptest.NewRequest("GET", "/orders", nil)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if tc.expectedStatusCode == http.StatusOK {
				var gotResponse []OrderResponse
				err := json.NewDecoder(rr.Body).Decode(&gotResponse)
				if err != nil {
					t.Errorf("failed to decode response: %v", err)
				}

				if len(gotResponse) != len(tc.expectedResponse) {
					t.Errorf("expected response length %d, got %d", len(tc.expectedResponse), len(gotResponse))
				}
				for i, expectedItem := range tc.expectedResponse {
					if gotResponse[i] != expectedItem {
						t.Errorf("expected response %v, got %v", expectedItem, gotResponse[i])
					}
				}
			}
		})
	}
}
