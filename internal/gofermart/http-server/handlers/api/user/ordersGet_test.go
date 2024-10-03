package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
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
	accrualMock := mocks.NewMockAccrualService(ctrl)
	mockUsernameExtractor := mocks.NewMockUsernameExtractor(ctrl)
	mockOrderRepo := mocks.NewMockOrderRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	testCases := []struct {
		name               string
		setupMocks         func()
		expectedStatusCode int
		expectedResponse   []OrderResponse
	}{
		{
			name: "Unauthorized_Access",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("", http.ErrNoCookie)
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "Successful_Order_Response",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrdersByUserID(gomock.Any(), gomock.Any()).Return([]domain.Order{
					{
						OrderNumber: "123",
						OrderStatus: domain.OrderStatusProcessed,
						Accrual:     150.5,
						UploadedAt:  time.Now(),
					},
				}, nil)
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
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrdersByUserID(gomock.Any(), gomock.Any()).Return([]domain.Order{}, nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "Internal_Server_Error",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), gomock.Any()).Return(&domain.User{UserID: 1}, nil)
				mockOrderRepo.EXPECT().GetOrdersByUserID(gomock.Any(), gomock.Any()).Return(nil, errors.New("database error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()
			mockOrderService := services.NewOrderService(accrualMock, mockOrderRepo, mockUserRepo, logger)

			handler := NewOrdersGetHandler(mockOrderService, mockUsernameExtractor, logger)

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
						diff := cmp.Diff(expectedItem, gotResponse[i])
						t.Errorf("expected response mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}
