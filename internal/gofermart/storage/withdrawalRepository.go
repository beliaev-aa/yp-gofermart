package storage

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WithdrawalRepository interface {
	AddWithdrawal(withdrawal domain.Withdrawal) error
	GetWithdrawalsByUserID(userID int) ([]domain.Withdrawal, error)
}

type WithdrawalRepositoryPostgres struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewWithdrawalRepository(db *gorm.DB, logger *zap.Logger) WithdrawalRepository {
	return &WithdrawalRepositoryPostgres{
		db:     db,
		logger: logger,
	}
}

// AddWithdrawal — добавление записи о выводе средств
func (w *WithdrawalRepositoryPostgres) AddWithdrawal(withdrawal domain.Withdrawal) error {
	w.logger.Info("Adding withdrawal", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.OrderNumber))
	err := w.db.Create(&withdrawal).Error
	if err != nil {
		w.logger.Error("Failed to add withdrawal", zap.Error(err))
		return err
	}
	w.logger.Info("Withdrawal added successfully", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.OrderNumber))
	return nil
}

// GetWithdrawalsByUserID — получение списка выводов пользователя
func (w *WithdrawalRepositoryPostgres) GetWithdrawalsByUserID(userID int) ([]domain.Withdrawal, error) {
	w.logger.Info("Getting withdrawals for user", zap.Int("userID", userID))
	var withdrawals []domain.Withdrawal
	err := w.db.Where("user_id = ?", userID).Order("processed_at desc").Find(&withdrawals).Error
	if err != nil {
		w.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}
	w.logger.Info("Withdrawals retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(withdrawals)))
	return withdrawals, nil
}
