package repository

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WithdrawalRepository interface {
	AddWithdrawal(tx *gorm.DB, withdrawal domain.Withdrawal) error
	GetWithdrawalsByUserID(tx *gorm.DB, userID int) ([]domain.Withdrawal, error)

	BeginTransaction() (*gorm.DB, error)
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
}

type WithdrawalRepositoryPostgres struct {
	*BaseRepository
}

func NewWithdrawalRepository(db *gorm.DB, logger *zap.Logger) WithdrawalRepository {
	return &WithdrawalRepositoryPostgres{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// AddWithdrawal — добавление записи о выводе средств
func (w *WithdrawalRepositoryPostgres) AddWithdrawal(tx *gorm.DB, withdrawal domain.Withdrawal) error {
	w.logger.Info("Adding withdrawal", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.OrderNumber))
	err := w.getDB(tx).Create(&withdrawal).Error
	if err != nil {
		w.logger.Error("Failed to add withdrawal", zap.Error(err))
		return err
	}
	w.logger.Info("Withdrawal added successfully", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.OrderNumber))
	return nil
}

// GetWithdrawalsByUserID — получение списка выводов пользователя
func (w *WithdrawalRepositoryPostgres) GetWithdrawalsByUserID(tx *gorm.DB, userID int) ([]domain.Withdrawal, error) {
	w.logger.Info("Getting withdrawals for user", zap.Int("userID", userID))
	var withdrawals []domain.Withdrawal
	err := w.getDB(tx).Where("user_id = ?", userID).Order("processed_at desc").Find(&withdrawals).Error
	if err != nil {
		w.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}
	w.logger.Info("Withdrawals retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(withdrawals)))
	return withdrawals, nil
}
