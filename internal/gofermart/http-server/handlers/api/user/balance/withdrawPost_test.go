package balance

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"bytes"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithdrawPostHandler_ServeHTTP(t *testing.T) {
	logger := zap.NewNop()

	type testCase struct {
		name               string
		requestBody        WithdrawPostRequest
		mockExtractFn      func(r *http.Request, logger *zap.Logger) (string, error)
		mockSetup          func(m *tests.MockStorage)
		expectedStatusCode int
		expectedResponse   string
	}

	testCases := []testCase{
		{
			name: "Unauthorized_Access",
			requestBody: WithdrawPostRequest{
				Order: "12345678903",
				Sum:   100.50,
			},
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "", http.ErrNoCookie
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Internal Server Error\n",
		},
		{
			name: "Invalid_Request_Format",
			requestBody: WithdrawPostRequest{
				Order: "",
				Sum:   0.0,
			},
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Internal Server Error\n",
		},
		{
			name: "Invalid_Order_Number_Format",
			requestBody: WithdrawPostRequest{
				Order: "123456",
				Sum:   100.50,
			},
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   "Invalid order number format\n",
		},
		{
			name: "Insufficient_Funds",
			requestBody: WithdrawPostRequest{
				Order: "79927398713",
				Sum:   150.00,
			},
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1}, nil
				}
			},
			expectedStatusCode: http.StatusPaymentRequired,
			expectedResponse:   "Insufficient funds\n",
		},
		{
			name: "Successful_Withdrawal",
			requestBody: WithdrawPostRequest{
				Order: "79927398713",
				Sum:   100.50,
			},
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1, Balance: 400}, nil
				}
				m.AddWithdrawalFn = func(withdrawal domain.Withdrawal) error {
					return nil
				}
				m.UpdateUserBalanceFn = func(userID int, amount float64) error {
					return nil
				}
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "",
		},
		{
			name: "Internal_Server_Error_On_Withdraw",
			requestBody: WithdrawPostRequest{
				Order: "79927398713",
				Sum:   100.50,
			},
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockSetup: func(m *tests.MockStorage) {
				m.GetUserByLoginFn = func(login string) (*domain.User, error) {
					return &domain.User{UserID: 1, Balance: 400}, nil
				}
				m.AddWithdrawalFn = func(withdrawal domain.Withdrawal) error {
					return errors.New("internal Server Error")
				}
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   "Internal Server Error\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExtractor := &tests.MockUsernameExtractor{
				ExtractFn: tc.mockExtractFn,
			}

			mockStorage := &tests.MockStorage{}
			if tc.mockSetup != nil {
				tc.mockSetup(mockStorage)
			}

			userService := services.NewUserService(mockStorage, logger)

			handler := NewWithdrawPostHandler(userService, mockExtractor, logger)

			body, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}
			req := httptest.NewRequest("POST", "/withdraw", bytes.NewBuffer(body))

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
