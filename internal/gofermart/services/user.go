package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"errors"
	"go.uber.org/zap"
	"time"
)

// UserService - отвечает за операции с пользователями, включая получение баланса и вывод средств
type UserService struct {
	storage storage.Storage
	logger  *zap.Logger
}

// NewUserService - создает новый экземпляр UserService
func NewUserService(storage storage.Storage, logger *zap.Logger) *UserService {
	return &UserService{
		storage: storage,
		logger:  logger,
	}
}

// GetBalance возвращает текущий баланс пользователя по его логину
func (s *UserService) GetBalance(login string) (*domain.UserBalance, error) {
	// Получаем баланс пользователя и сумму снятых средств из хранилища
	userBalance, err := s.storage.GetUserBalance(login)
	if err != nil {
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return nil, err
	}

	return userBalance, nil
}

// Withdraw обрабатывает запрос на вывод средств для указанного пользователя и заказа
func (s *UserService) Withdraw(login, order string, sum float64) error {
	// Получаем информацию о пользователе по логину
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		s.logger.Error("Failed to get user", zap.Error(err))
		return err
	}

	// Проверка на отрицательную сумму при выводе средств
	if sum <= 0 {
		return gofermartErrors.ErrInvalidWithdrawalAmount // Ошибка при попытке вывести недопустимую сумму
	}

	// Проверяем, достаточно ли средств для вывода
	if user.Balance < sum {
		return gofermartErrors.ErrInsufficientFunds // Ошибка недостатка средств
	}

	// Создаем запись о выводе средств
	withdrawal := domain.Withdrawal{
		OrderNumber: order,
		UserID:      user.UserID,
		Amount:      sum,
		ProcessedAt: time.Now(), // Время обработки вывода
	}

	// Добавляем информацию о выводе в хранилище
	err = s.storage.AddWithdrawal(withdrawal)
	if err != nil {
		s.logger.Error("Failed to add withdrawal", zap.Error(err))
		return err
	}

	// Обновляем баланс пользователя в хранилище
	err = s.storage.UpdateUserBalance(user.UserID, -sum) // Уменьшаем баланс
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err))
		return err
	}

	return nil
}

// GetWithdrawals возвращает список всех выводов средств пользователя по его логину
func (s *UserService) GetWithdrawals(login string) ([]domain.Withdrawal, error) {
	// Получаем информацию о пользователе по логину
	user, err := s.storage.GetUserByLogin(login)
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
	withdrawals, err := s.storage.GetWithdrawalsByUserID(user.UserID)
	if err != nil {
		s.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}

	return withdrawals, nil
}
