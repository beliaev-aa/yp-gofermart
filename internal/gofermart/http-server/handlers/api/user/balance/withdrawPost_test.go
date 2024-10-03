package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"bytes"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithdrawPostHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsernameExtractor := mocks.NewMockUsernameExtractor(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)

	logger := zap.NewNop()

	type testCase struct {
		name               string
		requestBody        string
		setupMocks         func()
		expectedStatusCode int
		expectedResponse   string
	}

	testCases := []testCase{
		{
			name:        "Malformed_JSON_Body",
			requestBody: `{"Order": "12345678903", "Sum": "invalid_number"}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "Invalid request format\n",
		},
		{
			name:        "Unauthorized_Access",
			requestBody: `{"Order": "12345678903", "Sum": 100.50}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("", http.ErrNoCookie)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Internal Server Error\n",
		},
		{
			name:        "Invalid_Request_Format",
			requestBody: `{"Order": "", "Sum": 0.0}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any()).Return(&domain.User{UserID: 1}, nil)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Internal Server Error\n",
		},
		{
			name:        "Invalid_Order_Number_Format",
			requestBody: `{"Order": "123456", "Sum": 100.50}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   "Invalid order number format\n",
		},
		{
			name:        "Insufficient_Funds",
			requestBody: `{"Order": "79927398713", "Sum": 150.00}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any()).Return(&domain.User{UserID: 1}, nil)
			},
			expectedStatusCode: http.StatusPaymentRequired,
			expectedResponse:   "Insufficient funds\n",
		},
		{
			name:        "Successful_Withdrawal",
			requestBody: `{"Order": "79927398713", "Sum": 100.50}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any()).Return(&domain.User{UserID: 1, Balance: 400}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "",
		},
		{
			name:        "Internal_Server_Error_On_Withdraw",
			requestBody: `{"Order": "79927398713", "Sum": 100.50}`,
			setupMocks: func() {
				mockUsernameExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any()).Return(&domain.User{UserID: 1, Balance: 400}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any()).Return(errors.New("internal Server Error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Internal Server Error\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			userService := services.NewUserService(mockUserRepo, mockWithdrawalRepo, logger)

			handler := NewWithdrawPostHandler(userService, mockUsernameExtractor, logger)

			req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(tc.requestBody))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if rr.Body.String() != tc.expectedResponse {
				t.Errorf("expected body %q, got %q", tc.expectedResponse, rr.Body.String())
			}
		})
	}
}
