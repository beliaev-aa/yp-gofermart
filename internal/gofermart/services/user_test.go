package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests"
	"errors"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestNewUserService(t *testing.T) {
	t.Run("NewUserService_CreatesService", func(t *testing.T) {
		mockStore := &tests.MockStorage{}
		logger := zap.NewNop()
		service := NewUserService(mockStore, logger)

		if service.storage != mockStore || service.logger != logger {
			t.Errorf("Expected UserService to be initialized with provided storage and logger")
		}
	})
}

func TestGetBalance(t *testing.T) {
	testCases := []struct {
		Name          string
		MockReturn    func() (float64, float64, error)
		ExpectedError error
		Expected      *Balance
	}{
		{
			Name: "GetBalance_Success",
			MockReturn: func() (float64, float64, error) {
				return 100, 50, nil
			},
			ExpectedError: nil,
			Expected:      &Balance{Current: 100, Withdrawn: 50},
		},
		{
			Name: "GetBalance_Failure",
			MockReturn: func() (float64, float64, error) {
				return 0, 0, errors.New("error")
			},
			ExpectedError: errors.New("error"),
			Expected:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockStore := &tests.MockStorage{
				GetUserBalanceFn: func(login string) (float64, float64, error) {
					return tc.MockReturn()
				},
			}
			logger := zap.NewNop()
			service := NewUserService(mockStore, logger)

			balance, err := service.GetBalance("test_user")

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			}

			if balance != nil && tc.Expected != nil {
				if balance.Current != tc.Expected.Current || balance.Withdrawn != tc.Expected.Withdrawn {
					t.Errorf("Expected balance %v, got %v", tc.Expected, balance)
				}
			}
		})
	}
}

func TestWithdraw(t *testing.T) {
	testCases := []struct {
		Name              string
		MockReturn        func() (*domain.User, error)
		MockAddWithdrawal func() error
		MockUpdateBalance func() error
		Sum               float64
		ExpectedError     error
	}{
		{
			Name: "Withdraw_Success",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1, Balance: 100}, nil
			},
			MockAddWithdrawal: func() error {
				return nil
			},
			MockUpdateBalance: func() error {
				return nil
			},
			Sum:           50,
			ExpectedError: nil,
		},
		{
			Name: "Withdraw_InsufficientFunds",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1, Balance: 10}, nil
			},
			MockAddWithdrawal: func() error {
				return nil
			},
			MockUpdateBalance: func() error {
				return nil
			},
			Sum:           50,
			ExpectedError: gofermartErrors.ErrInsufficientFunds,
		},
		{
			Name: "Withdraw_NegativeAmount",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1, Balance: 100}, nil
			},
			MockAddWithdrawal: func() error {
				return nil
			},
			MockUpdateBalance: func() error {
				return nil
			},
			Sum:           -10,
			ExpectedError: gofermartErrors.ErrInvalidWithdrawalAmount,
		},
		{
			Name: "Withdraw_FailedToGetUser",
			MockReturn: func() (*domain.User, error) {
				return nil, errors.New("failed to get user")
			},
			MockAddWithdrawal: func() error {
				return nil
			},
			MockUpdateBalance: func() error {
				return nil
			},
			Sum:           50,
			ExpectedError: errors.New("failed to get user"),
		},
		{
			Name: "Withdraw_FailedToAddWithdrawal",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1, Balance: 100}, nil
			},
			MockAddWithdrawal: func() error {
				return errors.New("failed to add withdrawal")
			},
			MockUpdateBalance: func() error {
				return nil
			},
			Sum:           50,
			ExpectedError: errors.New("failed to add withdrawal"),
		},
		{
			Name: "Withdraw_FailedToUpdateBalance",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1, Balance: 100}, nil
			},
			MockAddWithdrawal: func() error {
				return nil
			},
			MockUpdateBalance: func() error {
				return errors.New("failed to update balance")
			},
			Sum:           50,
			ExpectedError: errors.New("failed to update balance"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockStore := &tests.MockStorage{
				GetUserByLoginFn: func(login string) (*domain.User, error) {
					return tc.MockReturn()
				},
				AddWithdrawalFn: func(withdrawal domain.Withdrawal) error {
					return tc.MockAddWithdrawal()
				},
				UpdateUserBalanceFn: func(userID int, amount float64) error {
					return tc.MockUpdateBalance()
				},
			}
			logger := zap.NewNop()
			service := NewUserService(mockStore, logger)

			err := service.Withdraw("test_user", "order123", tc.Sum)

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && tc.ExpectedError != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.ExpectedError, err)
			}
		})
	}
}

func TestGetWithdrawals(t *testing.T) {
	testCases := []struct {
		Name            string
		MockReturn      func() (*domain.User, error)
		MockWithdrawals func() ([]domain.Withdrawal, error)
		ExpectedError   error
		Expected        []domain.Withdrawal
	}{
		{
			Name: "GetWithdrawals_Success",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1}, nil
			},
			MockWithdrawals: func() ([]domain.Withdrawal, error) {
				return []domain.Withdrawal{{OrderNumber: "order123", Amount: 50, ProcessedAt: time.Now()}}, nil
			},
			ExpectedError: nil,
			Expected:      []domain.Withdrawal{{OrderNumber: "order123", Amount: 50}},
		},
		{
			Name: "GetWithdrawals_UserNotFound",
			MockReturn: func() (*domain.User, error) {
				return nil, gofermartErrors.ErrUserNotFound
			},
			MockWithdrawals: func() ([]domain.Withdrawal, error) {
				return nil, nil
			},
			ExpectedError: gofermartErrors.ErrUserNotFound,
			Expected:      nil,
		},
		{
			Name: "GetWithdrawals_FailedToGetUser",
			MockReturn: func() (*domain.User, error) {
				return nil, errors.New("failed to get user")
			},
			MockWithdrawals: func() ([]domain.Withdrawal, error) {
				return nil, nil
			},
			ExpectedError: errors.New("failed to get user"),
			Expected:      nil,
		},
		{
			Name: "GetWithdrawals_FailedToGetWithdrawals",
			MockReturn: func() (*domain.User, error) {
				return &domain.User{UserID: 1}, nil
			},
			MockWithdrawals: func() ([]domain.Withdrawal, error) {
				return nil, errors.New("failed to get withdrawals")
			},
			ExpectedError: errors.New("failed to get withdrawals"),
			Expected:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockStore := &tests.MockStorage{
				GetUserByLoginFn: func(login string) (*domain.User, error) {
					return tc.MockReturn()
				},
				GetWithdrawalsByUserIDFn: func(userID int) ([]domain.Withdrawal, error) {
					return tc.MockWithdrawals()
				},
			}
			logger := zap.NewNop()
			service := NewUserService(mockStore, logger)

			withdrawals, err := service.GetWithdrawals("test_user")

			if err != nil && tc.ExpectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.ExpectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && tc.ExpectedError != nil && err.Error() != tc.ExpectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.ExpectedError, err)
			}

			if len(withdrawals) != len(tc.Expected) {
				t.Errorf("Expected %v withdrawals, got %v", len(tc.Expected), len(withdrawals))
			}
		})
	}
}
