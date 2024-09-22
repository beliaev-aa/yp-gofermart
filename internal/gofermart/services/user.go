package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"errors"
	"go.uber.org/zap"
	"time"
)

type UserService struct {
	storage storage.Storage
	logger  *zap.Logger
}

func NewUserService(storage storage.Storage, logger *zap.Logger) *UserService {
	return &UserService{
		storage: storage,
		logger:  logger,
	}
}

type Balance struct {
	Current   float64
	Withdrawn float64
}

func (s *UserService) GetBalance(login string) (*Balance, error) {
	userBalance, withdrawn, err := s.storage.GetUserBalance(login)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return nil, err
	}

	balance := &Balance{
		Current:   userBalance,
		Withdrawn: withdrawn,
	}

	return balance, nil
}

func (s *UserService) Withdraw(login, order string, sum float64) error {
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	if user.Balance < sum {
		return gofermartErrors.ErrInsufficientFunds
	}

	withdrawal := domain.Withdrawal{
		OrderNumber: order,
		UserID:      user.UserID,
		Amount:      sum,
		ProcessedAt: time.Now(),
	}

	err = s.storage.AddWithdrawal(withdrawal)
	if err != nil {
		s.logger.Error("Failed to add withdrawal", zap.Error(err))
		return err
	}

	err = s.storage.UpdateUserBalance(user.UserID, -sum)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err))
		return err
	}

	return nil
}

func (s *UserService) GetWithdrawals(login string) ([]domain.Withdrawal, error) {
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		if errors.Is(err, gofermartErrors.ErrUserNotFound) {
			s.logger.Warn("User not found", zap.String("login", login))
			return nil, gofermartErrors.ErrUserNotFound
		}
		s.logger.Error("Error getting user", zap.Error(err))
		return nil, err
	}

	withdrawals, err := s.storage.GetWithdrawalsByUserID(user.UserID)
	if err != nil {
		s.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}

	return withdrawals, nil
}
