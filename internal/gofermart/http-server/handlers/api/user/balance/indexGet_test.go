package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestIndexGetHandler_ServeHTTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsernameExtractor := mocks.NewMockUsernameExtractor(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)
	logger := zap.NewNop()

	testCases := []struct {
		name               string
		mockSetup          func(mockUserRepo *mocks.MockUserRepository, mockExtractor *mocks.MockUsernameExtractor)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "Unauthorized_Access",
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal Server Error\n",
			mockSetup: func(mockUserRepo *mocks.MockUserRepository, mockExtractor *mocks.MockUsernameExtractor) {
				mockExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("", http.ErrNoCookie)
			},
		},
		{
			name: "Internal_Server_Error_On_GetBalance",
			mockSetup: func(mockUserRepo *mocks.MockUserRepository, mockExtractor *mocks.MockUsernameExtractor) {
				mockExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed to get balance"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to get balance\n",
		},
		{
			name: "Successful_Balance_Response",
			mockSetup: func(mockUserRepo *mocks.MockUserRepository, mockExtractor *mocks.MockUsernameExtractor) {
				mockExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).Return("test_user", nil)
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), gomock.Any()).Return(&domain.UserBalance{
					Current:   100.50,
					Withdrawn: 50.75,
				}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"current":100.5,"withdrawn":50.75}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup(mockUserRepo, mockUsernameExtractor)
			userService := services.NewUserService(mockUserRepo, mockWithdrawalRepo, logger)

			handler := NewIndexGetHandler(userService, mockUsernameExtractor, logger)

			req := httptest.NewRequest("GET", "/balance", strings.NewReader(""))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tc.expectedBody) {
				diff := cmp.Diff(tc.expectedBody, rr.Body.String())
				t.Errorf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
