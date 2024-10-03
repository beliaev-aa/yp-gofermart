package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestUserService_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)
	logger := zap.NewNop()
	userService := NewUserService(mockUserRepo, mockWithdrawalRepo, logger)

	testCases := []struct {
		name           string
		login          string
		setupMocks     func()
		expectedError  error
		expectedResult *domain.UserBalance
	}{
		{
			name:  "GetBalance_Success",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserBalance("user1").Return(&domain.UserBalance{Current: 100.0}, nil)
			},
			expectedError:  nil,
			expectedResult: &domain.UserBalance{Current: 100.0},
		},
		{
			name:  "GetBalance_Error",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserBalance("user1").Return(nil, errors.New("db error"))
			},
			expectedError:  errors.New("db error"),
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			result, err := userService.GetBalance(tc.login)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}

			if result != nil && tc.expectedResult != nil && result.Current != tc.expectedResult.Current {
				t.Errorf("Expected result %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestUserService_Withdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)
	logger := zap.NewNop()
	userService := NewUserService(mockUserRepo, mockWithdrawalRepo, logger)

	testCases := []struct {
		name          string
		login         string
		order         string
		sum           float64
		setupMocks    func()
		expectedError error
	}{
		{
			name:  "Withdraw_Success",
			login: "user1",
			order: "order123",
			sum:   50.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(1, -50.0).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "Withdraw_Insufficient_Funds",
			login: "user1",
			order: "order123",
			sum:   150.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
			},
			expectedError: gofermartErrors.ErrInsufficientFunds,
		},
		{
			name:  "Withdraw_Invalid_Amount",
			login: "user1",
			order: "order123",
			sum:   -10.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
			},
			expectedError: gofermartErrors.ErrInvalidWithdrawalAmount,
		},
		{
			name:  "Withdraw_Add_Withdrawal_Error",
			login: "user1",
			order: "order123",
			sum:   50.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any()).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			err := userService.Withdraw(tc.login, tc.order, tc.sum)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestUserService_GetWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)
	logger := zap.NewNop()
	userService := NewUserService(mockUserRepo, mockWithdrawalRepo, logger)

	testCases := []struct {
		name           string
		login          string
		setupMocks     func()
		expectedError  error
		expectedResult []domain.Withdrawal
	}{
		{
			name:  "GetWithdrawals_Success",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(1).Return([]domain.Withdrawal{
					{OrderNumber: "order123", Amount: 50.0, ProcessedAt: time.Now()},
				}, nil)
			},
			expectedError: nil,
			expectedResult: []domain.Withdrawal{
				{OrderNumber: "order123", Amount: 50.0},
			},
		},
		{
			name:  "GetWithdrawals_User_Not_Found",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(nil, gofermartErrors.ErrUserNotFound)
			},
			expectedError:  gofermartErrors.ErrUserNotFound,
			expectedResult: nil,
		},
		{
			name:  "GetWithdrawals_Error",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin("user1").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(1).Return(nil, errors.New("db error"))
			},
			expectedError:  errors.New("db error"),
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			result, err := userService.GetWithdrawals(tc.login)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}

			if len(result) != len(tc.expectedResult) {
				t.Errorf("Expected %d results, got %d", len(tc.expectedResult), len(result))
			}
		})
	}
}
