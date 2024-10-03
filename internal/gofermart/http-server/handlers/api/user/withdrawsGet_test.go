package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWithdrawalsGetHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)
	mockUsernameExtractor := mocks.NewMockUsernameExtractor(ctrl)

	logger := zap.NewNop()
	userService := services.NewUserService(mockUserRepo, mockWithdrawalRepo, logger)
	handler := NewWithdrawalsGetHandler(userService, mockUsernameExtractor, logger)

	testCases := []struct {
		name                 string
		setupMocks           func()
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "ExtractUsername_Error",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("", errors.New("extraction error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Internal Server Error\n",
		},
		{
			name: "GetWithdrawals_Error",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return(nil, errors.New("db error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "Internal Server Error\n",
		},
		{
			name: "No_Content",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return([]domain.Withdrawal{}, nil)
			},
			expectedStatusCode:   http.StatusNoContent,
			expectedResponseBody: "",
		},
		{
			name: "Successful_Response",
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return([]domain.Withdrawal{
					{
						OrderNumber: "123456789",
						Amount:      decimal.NewFromFloat(100.50),
						ProcessedAt: time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC),
					},
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `[{"order":"123456789","sum":100.5,"processed_at":"2024-10-01T12:00:00Z"}]` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/withdrawals", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatusCode, rec.Code)
			}

			if strings.TrimSpace(rec.Body.String()) != strings.TrimSpace(tc.expectedResponseBody) {
				t.Errorf("Expected response body '%s', got '%s'", tc.expectedResponseBody, rec.Body.String())
			}
		})
	}
}
