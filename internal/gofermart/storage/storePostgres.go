//go:build !test

package storage

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// StorePostgres — структура для работы с базой данных PostgreSQL
type StorePostgres struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewStorage — создаёт новое хранилище с подключением к PostgreSQL и инициализирует схему базы данных
func NewStorage(dsn string, logger *zap.Logger) (*Storage, error) {
	// Подключение к базе данных через GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		return nil, err
	}

	store := &StorePostgres{
		db:     db,
		logger: logger,
	}

	// Инициализация схемы базы данных
	if err := store.initSchema(); err != nil {
		logger.Error("Failed to initialize database schema", zap.Error(err))
		return nil, err
	}

	return &Storage{
		UserRepo:       NewUserRepository(db, logger),
		OrderRepo:      NewOrderRepository(db, logger),
		WithdrawalRepo: NewWithdrawalRepository(db, logger),
	}, nil
}

// Close — закрывает подключение к базе данных
func (s *StorePostgres) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// initSchema — инициализация схемы базы данных с помощью миграций
func (s *StorePostgres) initSchema() error {
	// Автоматическая миграция схемы для пользователей, заказов и выводов
	return s.db.AutoMigrate(&domain.User{}, &domain.Order{}, &domain.Withdrawal{})
}
