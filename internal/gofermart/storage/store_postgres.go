package storage

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StorePostgres — структура для работы с базой данных PostgreSQL
type StorePostgres struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewStorage — создаёт новое хранилище с подключением к PostgreSQL и инициализирует схему базы данных
func NewStorage(dsn string, logger *zap.Logger) (Storage, error) {
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

	return store, nil
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

// GetUserByLogin — получение пользователя по логину
func (s *StorePostgres) GetUserByLogin(login string) (*domain.User, error) {
	s.logger.Info("Getting user by login", zap.String("login", login))
	var user domain.User
	err := s.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found", zap.String("login", login))
			return nil, gofermartErrors.ErrUserNotFound
		}
		s.logger.Error("Failed to get user by login", zap.Error(err))
		return nil, err
	}
	s.logger.Info("User retrieved successfully", zap.String("login", login))
	return &user, nil
}

// SaveUser — сохранение нового пользователя
func (s *StorePostgres) SaveUser(user domain.User) error {
	s.logger.Info("Saving new user", zap.String("login", user.Login))
	err := s.db.Create(&user).Error
	if err != nil {
		var pgErr *pq.Error
		// Проверка на ошибку уникальности (уникальный логин)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			s.logger.Warn("Login already exists", zap.String("login", user.Login))
			return gofermartErrors.ErrLoginAlreadyExists
		}
		s.logger.Error("Failed to save user", zap.Error(err))
		return err
	}
	s.logger.Info("User saved successfully", zap.String("login", user.Login))
	return nil
}

// UpdateUserBalance — обновление баланса пользователя
func (s *StorePostgres) UpdateUserBalance(userID int, amount float64) error {
	s.logger.Info("Updating user balance", zap.Int("userID", userID), zap.Float64("amount", amount))
	// Увеличение баланса пользователя
	err := s.db.Model(&domain.User{}).Where("user_id = ?", userID).Update("balance", gorm.Expr("balance + ?", amount)).Error
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Error(err))
		return err
	}
	s.logger.Info("User balance updated successfully", zap.Int("userID", userID))
	return nil
}

// AddWithdrawal — добавление записи о выводе средств
func (s *StorePostgres) AddWithdrawal(withdrawal domain.Withdrawal) error {
	s.logger.Info("Adding withdrawal", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.OrderNumber))
	err := s.db.Create(&withdrawal).Error
	if err != nil {
		s.logger.Error("Failed to add withdrawal", zap.Error(err))
		return err
	}
	s.logger.Info("Withdrawal added successfully", zap.Int("userID", withdrawal.UserID), zap.String("order", withdrawal.OrderNumber))
	return nil
}

// GetWithdrawalsByUserID — получение списка выводов пользователя
func (s *StorePostgres) GetWithdrawalsByUserID(userID int) ([]domain.Withdrawal, error) {
	s.logger.Info("Getting withdrawals for user", zap.Int("userID", userID))
	var withdrawals []domain.Withdrawal
	err := s.db.Where("user_id = ?", userID).Order("processed_at desc").Find(&withdrawals).Error
	if err != nil {
		s.logger.Error("Failed to get withdrawals", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Withdrawals retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(withdrawals)))
	return withdrawals, nil
}

// GetOrdersByUserID — получение списка заказов пользователя
func (s *StorePostgres) GetOrdersByUserID(userID int) ([]domain.Order, error) {
	s.logger.Info("Getting orders for user", zap.Int("userID", userID))
	var orders []domain.Order
	err := s.db.Where("user_id = ?", userID).Order("uploaded_at desc").Find(&orders).Error
	if err != nil {
		s.logger.Error("Failed to get orders", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Orders retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(orders)))
	return orders, nil
}

// AddOrder — добавление нового заказа
func (s *StorePostgres) AddOrder(order domain.Order) error {
	s.logger.Info("Adding new order", zap.String("order_number", order.OrderNumber), zap.Int("userID", order.UserID))

	// Использование OnConflict для игнорирования дублирующихся заказов
	err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&order).Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Проверка, является ли ошибка ошибкой Postgres
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Код ошибки уникального ограничения
				s.logger.Warn("Order number already exists", zap.String("order_number", order.OrderNumber))
				return gofermartErrors.ErrOrderAlreadyExists
			}
		}
		s.logger.Error("Failed to add order", zap.Error(err))
		return err
	}

	s.logger.Info("Order added successfully", zap.String("order_number", order.OrderNumber))
	return nil
}

// GetOrderByNumber — получение заказа по номеру
func (s *StorePostgres) GetOrderByNumber(number string) (*domain.Order, error) {
	s.logger.Info("Getting order by number", zap.String("order_number", number))
	var order domain.Order
	err := s.db.Where("order_number = ?", number).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Order not found", zap.String("order_number", number))
			return nil, nil
		}
		s.logger.Error("Failed to get order by number", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Order retrieved successfully", zap.String("order_number", number))
	return &order, nil
}

// UpdateOrder — обновление данных о заказе
func (s *StorePostgres) UpdateOrder(order domain.Order) error {
	s.logger.Info("Updating order", zap.String("order_number", order.OrderNumber))
	err := s.db.Model(&domain.Order{}).Where("order_number = ?", order.OrderNumber).Updates(domain.Order{
		OrderStatus: order.OrderStatus,
		Accrual:     order.Accrual,
	}).Error
	if err != nil {
		s.logger.Error("Failed to update order", zap.Error(err))
		return err
	}
	s.logger.Info("Order updated successfully", zap.String("order_number", order.OrderNumber))
	return nil
}

// GetOrdersForProcessing — получение заказов для обработки
func (s *StorePostgres) GetOrdersForProcessing() ([]domain.Order, error) {
	var orders []domain.Order
	// Получение заказов со статусом для обработки
	err := s.db.Where("order_status IN ? AND is_processing = ?", []string{
		domain.OrderStatusNew, domain.OrderStatusRegistered, domain.OrderStatusProcessing}, false).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&orders).Error
	if err != nil {
		s.logger.Error("Failed to get orders for processing", zap.Error(err))
		return nil, err
	}
	return orders, nil
}

// LockOrderForProcessing — блокировка заказа для обработки
func (s *StorePostgres) LockOrderForProcessing(orderNumber string) error {
	err := s.db.Model(&domain.Order{}).
		Where("order_number = ?", orderNumber).
		Update("is_processing", true).Error
	if err != nil {
		s.logger.Error("Failed to lock order for processing", zap.Error(err))
	}
	return err
}

// UnlockOrder — разблокировка заказа
func (s *StorePostgres) UnlockOrder(orderNumber string) error {
	err := s.db.Model(&domain.Order{}).
		Where("order_number = ?", orderNumber).
		Update("is_processing", false).Error
	if err != nil {
		s.logger.Error("Failed to unlock order", zap.Error(err))
	}
	return err
}

// GetUserBalance — получение баланса и общей суммы выводов пользователя
func (s *StorePostgres) GetUserBalance(login string) (userBalance *domain.UserBalance, err error) {
	var result *domain.UserBalance
	// Получение баланса пользователя и общей суммы выводов через Join
	err = s.db.Table("users").
		Select("users.balance AS current, COALESCE(SUM(withdrawals.amount), 0) AS withdrawn").
		Joins("LEFT JOIN withdrawals ON users.user_id = withdrawals.user_id").
		Where("users.login = ?", login).
		Group("users.balance").
		Scan(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gofermartErrors.ErrUserNotFound
		}
		s.logger.Error("Failed to get user balance", zap.Error(err))
		return nil, err
	}

	return result, nil
}
