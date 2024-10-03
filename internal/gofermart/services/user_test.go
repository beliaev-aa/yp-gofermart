package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/tests/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"
	"strings"
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
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), "user1").Return(&domain.UserBalance{Current: 100.0}, nil)
			},
			expectedError:  nil,
			expectedResult: &domain.UserBalance{Current: 100.0},
		},
		{
			name:  "GetBalance_Error",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserBalance(gomock.Any(), "user1").Return(nil, errors.New("db error"))
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
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockUserRepo.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 200.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "Withdraw_Insufficient_Funds",
			login: "user1",
			order: "order123",
			sum:   150.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
			},
			expectedError: gofermartErrors.ErrInsufficientFunds,
		},
		{
			name:  "Withdraw_Invalid_Amount",
			login: "user1",
			order: "order123",
			sum:   -10.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
			},
			expectedError: gofermartErrors.ErrInvalidWithdrawalAmount,
		},
		{
			name:  "Withdraw_Add_Withdrawal_Error",
			login: "user1",
			order: "order123",
			sum:   50.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
		{
			name:  "Withdraw_Fail_Get_User_By_Login",
			login: "user1",
			order: "order123",
			sum:   -10.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(nil, gofermartErrors.ErrUserNotFound)
			},
			expectedError: gofermartErrors.ErrUserNotFound,
		},
		{
			name:  "Withdraw_Fail_Update_User_Balance",
			login: "user1",
			order: "order123",
			sum:   50.0,
			setupMocks: func() {
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockUserRepo.EXPECT().Commit(gomock.Any()).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(nil).AnyTimes()
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 200.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("update user balance fail"))
			},
			expectedError: errors.New("update user balance fail"),
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
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return([]domain.Withdrawal{
					{OrderNumber: "order123", Amount: 50.0, ProcessedAt: time.Now()},
					{OrderNumber: "order456", Amount: 100.0, ProcessedAt: time.Now()},
				}, nil)
			},
			expectedError: nil,
			expectedResult: []domain.Withdrawal{
				{OrderNumber: "order123", Amount: 50.0},
				{OrderNumber: "order456", Amount: 100.0},
			},
		},
		{
			name:  "GetWithdrawals_User_Not_Found",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(nil, gofermartErrors.ErrUserNotFound)
			},
			expectedError:  gofermartErrors.ErrUserNotFound,
			expectedResult: nil,
		},
		{
			name:  "GetWithdrawals_DB_Error",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(nil, errors.New("db error"))
			},
			expectedError:  errors.New("db error"),
			expectedResult: nil,
		},
		{
			name:  "GetWithdrawals_Error_Fetching_Withdrawals",
			login: "user1",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1}, nil)
				mockWithdrawalRepo.EXPECT().GetWithdrawalsByUserID(gomock.Any(), 1).Return(nil, errors.New("db error"))
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
				t.Errorf("Expected result length %d, got %d", len(tc.expectedResult), len(result))
			} else {
				for i := range result {
					if result[i].OrderNumber != tc.expectedResult[i].OrderNumber || result[i].Amount != tc.expectedResult[i].Amount {
						t.Errorf("Expected result %v, got %v", tc.expectedResult, result)
					}
				}
			}
		})
	}
}

func TestUserService_Withdraw_TransactionHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockWithdrawalRepo := mocks.NewMockWithdrawalRepository(ctrl)

	core, observedLogs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	userService := NewUserService(mockUserRepo, mockWithdrawalRepo, logger)

	testCases := []struct {
		name          string
		setupMocks    func()
		expectedError error
		expectedLog   string
	}{
		{
			name: "Begin_Transaction_Error",
			setupMocks: func() {
				mockUserRepo.EXPECT().BeginTransaction().Return(nil, errors.New("failed to begin transaction"))
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
			},
			expectedError: errors.New("failed to begin transaction"),
			expectedLog:   "Failed to begin transaction",
		},
		{
			name: "Rollback_Failure_On_AddWithdrawal_Error",
			setupMocks: func() {
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(errors.New("add withdrawal error"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(errors.New("rollback error"))
			},
			expectedError: errors.New("add withdrawal error"),
			expectedLog:   "Failed to rollback transaction",
		},
		{
			name: "Rollback_Failure_On_UpdateUserBalance_Error",
			setupMocks: func() {
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 100.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("update user balance error"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(errors.New("rollback error"))
			},
			expectedError: errors.New("update user balance error"),
			expectedLog:   "Failed to rollback transaction",
		},
		{
			name: "Commit_Failure",
			setupMocks: func() {
				mockUserRepo.EXPECT().BeginTransaction().Return(&gorm.DB{}, nil)
				mockUserRepo.EXPECT().GetUserByLogin(gomock.Any(), "user1").Return(&domain.User{UserID: 1, Balance: 200.0}, nil)
				mockWithdrawalRepo.EXPECT().AddWithdrawal(gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().UpdateUserBalance(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockUserRepo.EXPECT().Commit(gomock.Any()).Return(errors.New("commit error"))
				mockUserRepo.EXPECT().Rollback(gomock.Any()).Return(errors.New("rollback error"))
			},
			expectedError: errors.New("commit error"),
			expectedLog:   "Failed to rollback transaction after failed commit",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMocks()

			err := userService.Withdraw("user1", "order123", 50.0)

			if err != nil && tc.expectedError == nil {
				t.Errorf("Expected no error, got %v", err)
			} else if err == nil && tc.expectedError != nil {
				t.Errorf("Expected error, got none")
			} else if err != nil && err.Error() != tc.expectedError.Error() {
				t.Errorf("Expected error %v, got %v", tc.expectedError, err)
			}

			found := false
			for _, log := range observedLogs.All() {
				if strings.Contains(log.Message, tc.expectedLog) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected log message not found: %v", tc.expectedLog)
			}
		})
	}
}
