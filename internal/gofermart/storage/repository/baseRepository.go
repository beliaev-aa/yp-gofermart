package repository

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type BaseRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewBaseRepository(db *gorm.DB, logger *zap.Logger) *BaseRepository {
	return &BaseRepository{
		db:     db,
		logger: logger,
	}
}

// Используем транзакцию, если она предоставлена, иначе используем обычное соединение
func (b *BaseRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return b.db
}

func (b *BaseRepository) BeginTransaction() (*gorm.DB, error) {
	tx := b.db.Begin()
	if tx.Error != nil {
		b.logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, tx.Error
	}
	return tx, nil
}

func (b *BaseRepository) Commit(tx *gorm.DB) error {
	if err := tx.Commit().Error; err != nil {
		b.logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}
	return nil
}

func (b *BaseRepository) Rollback(tx *gorm.DB) error {
	if err := tx.Rollback().Error; err != nil {
		b.logger.Error("Failed to rollback transaction", zap.Error(err))
		return err
	}
	return nil
}
