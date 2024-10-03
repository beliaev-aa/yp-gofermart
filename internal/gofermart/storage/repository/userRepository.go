package repository

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"errors"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserBalance(tx *gorm.DB, login string) (*domain.UserBalance, error)
	GetUserByLogin(tx *gorm.DB, login string) (*domain.User, error)
	SaveUser(tx *gorm.DB, user domain.User) error
	UpdateUserBalance(tx *gorm.DB, userID int, amount float64) error

	BeginTransaction() (*gorm.DB, error)
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
}

type UserRepositoryPostgres struct {
	*BaseRepository
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &UserRepositoryPostgres{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// GetUserBalance — получение баланса и общей суммы выводов пользователя
func (u *UserRepositoryPostgres) GetUserBalance(tx *gorm.DB, login string) (userBalance *domain.UserBalance, err error) {
	var result *domain.UserBalance
	// Получение баланса пользователя и общей суммы выводов через Join
	err = u.getDB(tx).Table("users").
		Select("users.balance AS current, COALESCE(SUM(withdrawals.amount), 0) AS withdrawn").
		Joins("LEFT JOIN withdrawals ON users.user_id = withdrawals.user_id").
		Where("users.login = ?", login).
		Group("users.balance").
		Scan(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gofermartErrors.ErrUserNotFound
		}
		u.logger.Error("Failed to get user balance", zap.Error(err))
		return nil, err
	}

	return result, nil
}

// GetUserByLogin — получение пользователя по логину
func (u *UserRepositoryPostgres) GetUserByLogin(tx *gorm.DB, login string) (*domain.User, error) {
	u.logger.Info("Getting user by login", zap.String("login", login))
	var user domain.User
	err := u.getDB(tx).Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			u.logger.Warn("User not found", zap.String("login", login))
			return nil, gofermartErrors.ErrUserNotFound
		}
		u.logger.Error("Failed to get user by login", zap.Error(err))
		return nil, err
	}
	u.logger.Info("User retrieved successfully", zap.String("login", login))
	return &user, nil
}

// SaveUser — сохранение нового пользователя
func (u *UserRepositoryPostgres) SaveUser(tx *gorm.DB, user domain.User) error {
	u.logger.Info("Saving new user", zap.String("login", user.Login))
	err := u.getDB(tx).Create(&user).Error
	if err != nil {
		var pgErr *pq.Error
		// Проверка на ошибку уникальности (уникальный логин)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			u.logger.Warn("Login already exists", zap.String("login", user.Login))
			return gofermartErrors.ErrLoginAlreadyExists
		}
		u.logger.Error("Failed to save user", zap.Error(err))
		return err
	}
	u.logger.Info("User saved successfully", zap.String("login", user.Login))
	return nil
}

// UpdateUserBalance — обновление баланса пользователя
func (u *UserRepositoryPostgres) UpdateUserBalance(tx *gorm.DB, userID int, amount float64) error {
	u.logger.Info("Updating user balance", zap.Int("userID", userID), zap.Float64("amount", amount))
	// Увеличение баланса пользователя
	err := u.getDB(tx).Model(&domain.User{}).Where("user_id = ?", userID).Update("balance", gorm.Expr("balance + ?", amount)).Error
	if err != nil {
		u.logger.Error("Failed to update user balance", zap.Error(err))
		return err
	}
	u.logger.Info("User balance updated successfully", zap.Int("userID", userID))
	return nil
}
