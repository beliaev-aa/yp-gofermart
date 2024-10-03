package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage/repository"
	"context"
	"errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// OrderServiceInterface - интерфейс для сервиса работы с заказами.
type OrderServiceInterface interface {
	AddOrder(login, number string) error
	GetOrders(login string) ([]domain.Order, error)
	UpdateOrderStatuses(ctx context.Context)
	UpdateUserBalance(tx *gorm.DB, userID int, amount float64) error
}

// OrderService - представляет сервис для работы с заказами.
type OrderService struct {
	accrualClient AccrualService
	logger        *zap.Logger
	orderRepo     repository.OrderRepository
	userRepo      repository.UserRepository
}

// NewOrderService - создает новый экземпляр OrderService.
func NewOrderService(accrualClient AccrualService, orderRepo repository.OrderRepository, userRepo repository.UserRepository, logger *zap.Logger) *OrderService {
	return &OrderService{
		accrualClient: accrualClient,
		logger:        logger,
		orderRepo:     orderRepo,
		userRepo:      userRepo,
	}
}

// AddOrder - добавляет новый заказ, проверяя, не был ли он уже добавлен другим пользователем.
func (s *OrderService) AddOrder(login, number string) error {
	// Получаем пользователя по логину
	user, err := s.userRepo.GetUserByLogin(nil, login)
	if err != nil {
		return err
	}

	// Проверяем, был ли уже добавлен заказ с таким номером
	existingOrder, err := s.orderRepo.GetOrderByNumber(nil, number)
	if err != nil && !errors.Is(err, gofermartErrors.ErrOrderNotFound) {
		return err
	}

	if existingOrder != nil {
		// Если заказ добавлен текущим пользователем
		if existingOrder.UserID == user.UserID {
			return gofermartErrors.ErrOrderAlreadyUploaded
		}
		// Если заказ добавлен другим пользователем
		return gofermartErrors.ErrOrderUploadedByAnother
	}

	// Создаем новый заказ и добавляем его в хранилище
	order := domain.Order{
		OrderNumber: number,
		UserID:      user.UserID,
		OrderStatus: domain.OrderStatusNew,
		UploadedAt:  time.Now(),
	}

	err = s.orderRepo.AddOrder(nil, order)
	if err != nil {
		return err
	}

	return nil
}

// GetOrders - возвращает список заказов пользователя.
func (s *OrderService) GetOrders(login string) ([]domain.Order, error) {
	user, err := s.userRepo.GetUserByLogin(nil, login)
	if err != nil {
		return nil, err
	}

	orders, err := s.orderRepo.GetOrdersByUserID(nil, user.UserID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// UpdateOrderStatuses - обновляет статусы заказов путем запроса к внешней системе начисления.
func (s *OrderService) UpdateOrderStatuses(ctx context.Context) {
	s.logger.Info("Starting order status update")

	// Получаем заказы, которые не обрабатываются (processing = FALSE)
	orders, err := s.orderRepo.GetOrdersForProcessing(nil)
	if err != nil {
		s.logger.Error("Failed to fetch orders for status update", zap.Error(err))
		return
	}

	for _, order := range orders {
		// Начинаем новую транзакцию для обработки каждого заказа
		tx, err := s.userRepo.BeginTransaction()
		if err != nil {
			s.logger.Error("Failed to start transaction", zap.Error(err))
			continue
		}

		// Блокируем заказ для обработки, устанавливаем processing = TRUE
		err = s.orderRepo.LockOrderForProcessing(tx, order.OrderNumber)
		if err != nil {
			s.logger.Error("Failed to lock order for processing", zap.String("order", order.OrderNumber), zap.Error(err))
			if err := s.userRepo.Rollback(tx); err != nil {
				s.logger.Error("Failed to rollback transaction", zap.Error(err))
			}
			continue
		}

		success := s.processOrder(ctx, order, tx)

		if success {
			if err := s.userRepo.Commit(tx); err != nil {
				s.logger.Error("Failed to commit transaction", zap.Error(err))
			}
		} else {
			if err := s.userRepo.Rollback(tx); err != nil {
				s.logger.Error("Failed to rollback transaction", zap.Error(err))
			}
		}
	}
}

// processOrder - обрабатывает конкретный заказ в рамках транзакции
func (s *OrderService) processOrder(ctx context.Context, order domain.Order, tx *gorm.DB) bool {
	accrual, status, err := s.accrualClient.GetOrderAccrual(ctx, order.OrderNumber)
	if err != nil {
		s.logger.Warn("Failed to fetch order accrual", zap.String("order", order.OrderNumber), zap.Error(err))
		return false
	}

	order.OrderStatus = status
	order.Accrual = accrual

	if err := s.orderRepo.UpdateOrder(tx, order); err != nil {
		s.logger.Error("Failed to update order", zap.String("order", order.OrderNumber), zap.Error(err))
		return false
	}

	if status == domain.OrderStatusProcessed {
		if err := s.UpdateUserBalance(tx, order.UserID, accrual); err != nil {
			s.logger.Error("Failed to update user balance", zap.Int("userID", order.UserID), zap.Error(err))
			return false
		}
	}

	// Снимаем блокировку после завершения обработки заказа (processing = FALSE)
	if err := s.orderRepo.UnlockOrder(tx, order.OrderNumber); err != nil {
		s.logger.Error("Failed to unlock order", zap.String("order", order.OrderNumber), zap.Error(err))
		return false
	}

	s.logger.Info("Order processed successfully", zap.String("order", order.OrderNumber))

	return true
}

// UpdateUserBalance - обновляет баланс пользователя.
func (s *OrderService) UpdateUserBalance(tx *gorm.DB, userID int, amount float64) error {
	err := s.userRepo.UpdateUserBalance(tx, userID, amount)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Int("userID", userID), zap.Error(err))
		return err
	}
	return nil
}
