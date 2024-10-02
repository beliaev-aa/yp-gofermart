package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
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
	mockExtractor := mocks.NewMockUsernameExtractor(ctrl)
	logger := zap.NewNop()

	testCases := []struct {
		name               string
		mockExtractFn      func(r *http.Request, logger *zap.Logger) (string, error)
		mockSetup          func(m *tests.MockStorage)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Unauthorized_Access",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "", http.ErrNoCookie
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Internal Server Error\n",
		},
		{
			name: "Internal_Server_Error_On_GetBalance",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.GetUserBalanceFn = func(login string) (userBalance *domain.UserBalance, err error) {
					return nil, errors.New("failed to get balance")
				}
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "Failed to get balance\n",
		},
		{
			name: "Successful_Balance_Response",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.GetUserBalanceFn = func(login string) (userBalance *domain.UserBalance, err error) {
					return &domain.UserBalance{
						Current:   100.50,
						Withdrawn: 50.75,
					}, nil
				}
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"current":100.5,"withdrawn":50.75}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExtractor.EXPECT().ExtractUsernameFromContext(gomock.Any(), gomock.Any()).DoAndReturn(tc.mockExtractFn)

			mockStorage := &tests.MockStorage{}
			if tc.mockSetup != nil {
				tc.mockSetup(mockStorage)
			}

			userService := services.NewUserService(mockStorage, logger)

			handler := NewIndexGetHandler(userService, mockExtractor, logger)

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
