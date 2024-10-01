package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/tests"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithdrawalsGetHandler_ServeHTTP(t *testing.T) {
	logger := zap.NewNop()

	testWithdrawals := []domain.Withdrawal{
		{
			OrderNumber: "123456",
			Amount:      100.50,
			ProcessedAt: time.Now(),
		},
	}

	testUser := &domain.User{
		UserID: 1,
		Login:  "test_user",
	}

	testCases := []struct {
		name               string
		mockExtractFn      func(r *http.Request, logger *zap.Logger) (string, error)
		withdrawals        []domain.Withdrawal
		mockError          error
		expectedStatusCode int
		expectedResponse   []WithdrawalResponse
	}{
		{
			name: "Success_Response_With_Data",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			withdrawals:        testWithdrawals,
			expectedStatusCode: http.StatusOK,
			expectedResponse: []WithdrawalResponse{
				{
					Order:       "123456",
					Sum:         100.50,
					ProcessedAt: testWithdrawals[0].ProcessedAt.Format(time.RFC3339),
				},
			},
		},
		{
			name: "Unauthorized_Access",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "", http.ErrNoCookie
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "No_Content",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			withdrawals:        []domain.Withdrawal{},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "Service_Error",
			mockExtractFn: func(r *http.Request, logger *zap.Logger) (string, error) {
				return "test_user", nil
			},
			mockError:          errors.New("internal service error"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExtractor := &tests.MockUsernameExtractor{
				ExtractFn: tc.mockExtractFn,
			}

			mockUserService := services.NewUserService(&tests.MockStorage{
				GetUserByLoginFn: func(login string) (*domain.User, error) {
					if login == "test_user" {
						return testUser, nil
					}
					return nil, errors.New("user not found")
				},
				GetWithdrawalsByUserIDFn: func(userID int) ([]domain.Withdrawal, error) {
					return tc.withdrawals, tc.mockError
				},
			}, logger)

			handler := NewWithdrawalsGetHandler(mockUserService, mockExtractor, logger)

			req := httptest.NewRequest("GET", "/withdrawals", nil)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatusCode {
				t.Errorf("expected status %v, got %v", tc.expectedStatusCode, rr.Code)
			}

			if tc.expectedStatusCode == http.StatusOK {
				var gotResponse []WithdrawalResponse
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
