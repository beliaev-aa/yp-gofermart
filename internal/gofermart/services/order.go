package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage/repository"
	"context"
	"errors"
	"go.uber.org/zap"
	"time"
)

// OrderServiceInterface - интерфейс для сервиса работы с заказами.
type OrderServiceInterface interface {
	AddOrder(login, number string) error
	GetOrders(login string) ([]domain.Order, error)
	UpdateOrderStatuses(ctx context.Context)
	UpdateUserBalance(userID int, amount float64) error
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
	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return err
	}

	// Проверяем, был ли уже добавлен заказ с таким номером
	existingOrder, err := s.orderRepo.GetOrderByNumber(number)
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

	err = s.orderRepo.AddOrder(order)
	if err != nil {
		return err
	}

	return nil
}

// GetOrders - возвращает список заказов пользователя.
func (s *OrderService) GetOrders(login string) ([]domain.Order, error) {
	user, err := s.userRepo.GetUserByLogin(login)
	if err != nil {
		return nil, err
	}

	orders, err := s.orderRepo.GetOrdersByUserID(user.UserID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// UpdateOrderStatuses - обновляет статусы заказов путем запроса к внешней системе начисления.
func (s *OrderService) UpdateOrderStatuses(ctx context.Context) {
	s.logger.Info("Starting order status update")

	// Получаем заказы, которые не обрабатываются (processing = FALSE)
	orders, err := s.orderRepo.GetOrdersForProcessing()
	if err != nil {
		s.logger.Error("Failed to fetch orders for status update", zap.Error(err))
		return
	}

	for _, order := range orders {
		// Блокируем заказ для обработки, устанавливаем processing = TRUE
		err := s.orderRepo.LockOrderForProcessing(order.OrderNumber)
		if err != nil {
			s.logger.Error("Failed to lock order for processing", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}

		// Обработка заказа
		s.processOrder(ctx, order)

		// Снимаем блокировку после завершения обработки заказа (processing = FALSE)
		err = s.orderRepo.UnlockOrder(order.OrderNumber)
		if err != nil {
			s.logger.Error("Failed to unlock order", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}
	}
}

// processOrder - обрабатывает конкретный заказ
func (s *OrderService) processOrder(ctx context.Context, order domain.Order) {
	accrual, status, err := s.accrualClient.GetOrderAccrual(ctx, order.OrderNumber)
	if err != nil {
		s.logger.Warn("Failed to fetch order accrual", zap.String("order", order.OrderNumber), zap.Error(err))
		return
	}

	order.OrderStatus = status
	order.Accrual = accrual

	err = s.orderRepo.UpdateOrder(order)
	if err != nil {
		s.logger.Error("Failed to update order", zap.String("order", order.OrderNumber), zap.Error(err))
		return
	}

	if status == domain.OrderStatusProcessed {
		err = s.UpdateUserBalance(order.UserID, accrual)
		if err != nil {
			s.logger.Error("Failed to update user balance", zap.Int("userID", order.UserID), zap.Error(err))
			return
		}
	}
}

// UpdateUserBalance - обновляет баланс пользователя.
func (s *OrderService) UpdateUserBalance(userID int, amount float64) error {
	err := s.userRepo.UpdateUserBalance(userID, amount)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Int("userID", userID), zap.Error(err))
		return err
	}
	return nil
}
