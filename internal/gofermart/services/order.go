package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"errors"
	"go.uber.org/zap"
	"time"
)

// OrderService - представляет сервис для работы с заказами.
type OrderService struct {
	accrualClient AccrualService
	logger        *zap.Logger
	storage       storage.Storage
}

// NewOrderService - создает новый экземпляр OrderService.
func NewOrderService(accrualClient AccrualService, storage storage.Storage, logger *zap.Logger) *OrderService {
	return &OrderService{
		accrualClient: accrualClient,
		logger:        logger,
		storage:       storage,
	}
}

// AddOrder - добавляет новый заказ, проверяя, не был ли он уже добавлен другим пользователем.
func (s *OrderService) AddOrder(login, number string) error {
	// Получаем пользователя по логину
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		return err
	}

	// Проверяем, был ли уже добавлен заказ с таким номером
	existingOrder, err := s.storage.GetOrderByNumber(number)
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

	err = s.storage.AddOrder(order)
	if err != nil {
		return err
	}

	return nil
}

// GetOrders - возвращает список заказов пользователя.
func (s *OrderService) GetOrders(login string) ([]domain.Order, error) {
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		return nil, err
	}

	orders, err := s.storage.GetOrdersByUserID(user.UserID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// UpdateOrderStatuses - обновляет статусы заказов путем запроса к внешней системе начисления.
func (s *OrderService) UpdateOrderStatuses() {
	s.logger.Info("Starting order status update")

	// Получаем заказы, которые не обрабатываются (processing = FALSE)
	orders, err := s.storage.GetOrdersForProcessing()
	if err != nil {
		s.logger.Error("Failed to fetch orders for status update", zap.Error(err))
		return
	}

	for _, order := range orders {
		// Блокируем заказ для обработки, устанавливаем processing = TRUE
		err := s.storage.LockOrderForProcessing(order.OrderNumber)
		if err != nil {
			s.logger.Error("Failed to lock order for processing", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}

		// Обработка заказа
		s.processOrder(order)

		// Снимаем блокировку после завершения обработки заказа (processing = FALSE)
		err = s.storage.UnlockOrder(order.OrderNumber)
		if err != nil {
			s.logger.Error("Failed to unlock order", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}
	}
}

// processOrder - обрабатывает конкретный заказ
func (s *OrderService) processOrder(order domain.Order) {
	accrual, status, err := s.accrualClient.GetOrderAccrual(order.OrderNumber)
	if err != nil {
		s.logger.Warn("Failed to fetch order accrual", zap.String("order", order.OrderNumber), zap.Error(err))
		return
	}

	order.OrderStatus = status
	order.Accrual = accrual

	err = s.storage.UpdateOrder(order)
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
	err := s.storage.UpdateUserBalance(userID, amount)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Int("userID", userID), zap.Error(err))
		return err
	}
	return nil
}
