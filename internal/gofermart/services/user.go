package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage/repository"
	"errors"
	"go.uber.org/zap"
	"time"
)

// UserService - отвечает за операции с пользователями, включая получение баланса и вывод средств
type UserService struct {
	logger         *zap.Logger
	userRepo       repository.UserRepository
	withdrawalRepo repository.WithdrawalRepository
}

// NewUserService - создает новый экземпляр UserService
func NewUserService(userRepo repository.UserRepository, withdrawalRepo repository.WithdrawalRepository, logger *zap.Logger) *UserService {
	return &UserService{
		logger:         logger,
		userRepo:       userRepo,
		withdrawalRepo: withdrawalRepo,
	}
}

// GetBalance возвращает текущий баланс пользователя по его логину
func (s *UserService) GetBalance(login string) (*domain.UserBalance, error) {
	// Получаем баланс пользователя и сумму снятых средств из хранилища
	userBalance, err := s.userRepo.GetUserBalance(nil, login)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return nil, err
	}

	return userBalance, nil
}

// Withdraw обрабатывает запрос на вывод средств для указанного пользователя и заказа
func (s *UserService) Withdraw(login, order string, sum float64) error {
	// Получаем информацию о пользователе по логину
	user, err := s.userRepo.GetUserByLogin(nil, login)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	// Проверка на отрицательную сумму при выводе средств
	if sum <= 0 {
		return gofermartErrors.ErrInvalidWithdrawalAmount
	}

	// Проверяем, достаточно ли средств для вывода
	if user.Balance < sum {
		return gofermartErrors.ErrInsufficientFunds
	}

	// Создаем запись о выводе средств
	withdrawal := domain.Withdrawal{
		OrderNumber: order,
		UserID:      user.UserID,
		Amount:      sum,
		ProcessedAt: time.Now(),
	}

	tx, err := s.userRepo.BeginTransaction()
	if err != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}

	// Добавляем информацию о выводе в хранилище
	err = s.withdrawalRepo.AddWithdrawal(tx, withdrawal)
	if err != nil {
		s.logger.Error("Failed to add withdrawal", zap.Error(err))
		if rbErr := s.userRepo.Rollback(tx); rbErr != nil {
			s.logger.Error("Failed to rollback transaction", zap.Error(rbErr))
		}
		return err
	}

	err = s.userRepo.UpdateUserBalance(tx, user.UserID, -sum)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err))
		if rbErr := s.userRepo.Rollback(tx); rbErr != nil {
			s.logger.Error("Failed to rollback transaction", zap.Error(rbErr))
		}
		return err
	}

	err = s.userRepo.Commit(tx)
	if err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		if rbErr := s.userRepo.Rollback(tx); rbErr != nil {
			s.logger.Error("Failed to rollback transaction after failed commit", zap.Error(rbErr))
		}
		return err
	}

	return nil
}

// GetWithdrawals возвращает список всех выводов средств пользователя по его логину
func (s *UserService) GetWithdrawals(login string) ([]domain.Withdrawal, error) {
	// Получаем информацию о пользователе по логину
	user, err := s.userRepo.GetUserByLogin(nil, login)
	if err != nil {
		// Проверяем, если пользователь не найден, возвращаем соответствующую ошибку
		if errors.Is(err, gofermartErrors.ErrUserNotFound) {
			s.logger.Warn("User not found", zap.String("login", login))
			return nil, gofermartErrors.ErrUserNotFound
		}
		s.logger.Error("Error getting user", zap.Error(err))
		return nil, err
	}

	// Получаем список всех выводов средств пользователя
	withdrawals, err := s.withdrawalRepo.GetWithdrawalsByUserID(nil, user.UserID)
	if err != nil {
		s.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}

	return withdrawals, nil
}
