package repository

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	AddOrder(tx *gorm.DB, order domain.Order) error
	GetOrderByNumber(tx *gorm.DB, number string) (*domain.Order, error)
	GetOrdersByUserID(tx *gorm.DB, userID int) ([]domain.Order, error)
	GetOrdersForProcessing(tx *gorm.DB) ([]domain.Order, error)
	LockOrderForProcessing(tx *gorm.DB, orderNumber string) error
	UnlockOrder(tx *gorm.DB, orderNumber string) error
	UpdateOrder(tx *gorm.DB, order domain.Order) error

	BeginTransaction() (*gorm.DB, error)
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB) error
}

type OrderRepositoryPostgres struct {
	*BaseRepository
}

func NewOrderRepository(db *gorm.DB, logger *zap.Logger) OrderRepository {
	return &OrderRepositoryPostgres{
		BaseRepository: NewBaseRepository(db, logger),
	}
}

// AddOrder — добавление нового заказа
func (o *OrderRepositoryPostgres) AddOrder(tx *gorm.DB, order domain.Order) error {
	o.logger.Info("Adding new order", zap.String("order_number", order.OrderNumber), zap.Int("userID", order.UserID))

	// Использование OnConflict для игнорирования дублирующихся заказов
	err := o.getDB(tx).Clauses(clause.OnConflict{DoNothing: true}).Create(&order).Error
	if err != nil {
		var pgErr *pgconn.PgError
		// Проверка, является ли ошибка ошибкой Postgres
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Код ошибки уникального ограничения
				o.logger.Warn("Order number already exists", zap.String("order_number", order.OrderNumber))
				return gofermartErrors.ErrOrderAlreadyExists
			}
		}
		o.logger.Error("Failed to add order", zap.Error(err))
		return err
	}

	o.logger.Info("Order added successfully", zap.String("order_number", order.OrderNumber))
	return nil
}

// GetOrderByNumber — получение заказа по номеру
func (o *OrderRepositoryPostgres) GetOrderByNumber(tx *gorm.DB, number string) (*domain.Order, error) {
	o.logger.Info("Getting order by number", zap.String("order_number", number))
	var order domain.Order
	err := o.getDB(tx).Where("order_number = ?", number).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			o.logger.Warn("Order not found", zap.String("order_number", number))
			return nil, nil
		}
		o.logger.Error("Failed to get order by number", zap.Error(err))
		return nil, err
	}
	o.logger.Info("Order retrieved successfully", zap.String("order_number", number))
	return &order, nil
}

// GetOrdersByUserID — получение списка заказов пользователя
func (o *OrderRepositoryPostgres) GetOrdersByUserID(tx *gorm.DB, userID int) ([]domain.Order, error) {
	o.logger.Info("Getting orders for user", zap.Int("userID", userID))
	var orders []domain.Order
	err := o.getDB(tx).Where("user_id = ?", userID).Order("uploaded_at desc").Find(&orders).Error
	if err != nil {
		o.logger.Error("Failed to get orders", zap.Error(err))
		return nil, err
	}
	o.logger.Info("Orders retrieved successfully", zap.Int("userID", userID), zap.Int("count", len(orders)))
	return orders, nil
}

// GetOrdersForProcessing — получение заказов для обработки
func (o *OrderRepositoryPostgres) GetOrdersForProcessing(tx *gorm.DB) ([]domain.Order, error) {
	var orders []domain.Order
	// Получение заказов со статусом для обработки
	err := o.getDB(tx).Where("order_status IN ? AND is_processing = ?", []string{
		domain.OrderStatusNew, domain.OrderStatusRegistered, domain.OrderStatusProcessing}, false).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&orders).Error
	if err != nil {
		o.logger.Error("Failed to get orders for processing", zap.Error(err))
		return nil, err
	}
	return orders, nil
}

// LockOrderForProcessing — блокировка заказа для обработки
func (o *OrderRepositoryPostgres) LockOrderForProcessing(tx *gorm.DB, orderNumber string) error {
	err := o.getDB(tx).Model(&domain.Order{}).
		Where("order_number = ?", orderNumber).
		Update("is_processing", true).Error
	if err != nil {
		o.logger.Error("Failed to lock order for processing", zap.Error(err))
	}
	return err
}

// UnlockOrder — разблокировка заказа
func (o *OrderRepositoryPostgres) UnlockOrder(tx *gorm.DB, orderNumber string) error {
	err := o.getDB(tx).Model(&domain.Order{}).
		Where("order_number = ?", orderNumber).
		Update("is_processing", false).Error
	if err != nil {
		o.logger.Error("Failed to unlock order", zap.Error(err))
	}
	return err
}

// UpdateOrder — обновление данных о заказе
func (o *OrderRepositoryPostgres) UpdateOrder(tx *gorm.DB, order domain.Order) error {
	o.logger.Info("Updating order", zap.String("order_number", order.OrderNumber))
	err := o.getDB(tx).Model(&domain.Order{}).Where("order_number = ?", order.OrderNumber).Updates(domain.Order{
		OrderStatus: order.OrderStatus,
		Accrual:     order.Accrual,
	}).Error
	if err != nil {
		o.logger.Error("Failed to update order", zap.Error(err))
		return err
	}
	o.logger.Info("Order updated successfully", zap.String("order_number", order.OrderNumber))
	return nil
}
