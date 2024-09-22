package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/storage"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type OrderService struct {
	storage    storage.Storage
	logger     *zap.Logger
	accrualURL string
}

func NewOrderService(accrualURL string, storage storage.Storage, logger *zap.Logger) *OrderService {
	return &OrderService{
		storage:    storage,
		logger:     logger,
		accrualURL: accrualURL,
	}
}

func (s *OrderService) AddOrder(login, number string) error {
	user, err := s.storage.GetUserByLogin(login)
	if err != nil {
		return err
	}

	existingOrder, err := s.storage.GetOrderByNumber(number)
	if err != nil && !errors.Is(err, gofermartErrors.ErrOrderNotFound) {
		return err
	}

	if existingOrder != nil {
		if existingOrder.UserID == user.UserID {
			return gofermartErrors.ErrOrderAlreadyUploaded
		} else {
			return gofermartErrors.ErrOrderUploadedByAnother
		}
	}

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

		// Обращаемся к внешней системе начисления
		accrual, status, err := s.fetchOrderAccrual(order.OrderNumber)
		if err != nil {
			s.logger.Warn("Failed to fetch order accrual", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}

		// Обновляем статус и начисление заказа
		order.OrderStatus = status
		order.Accrual = accrual

		err = s.storage.UpdateOrder(order)
		if err != nil {
			s.logger.Error("Failed to update order", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}

		// Если заказ обработан, обновляем баланс пользователя
		if status == domain.OrderStatusProcessed {
			err = s.UpdateUserBalance(order.UserID, accrual)
			if err != nil {
				s.logger.Error("Failed to update user balance", zap.Int("userID", order.UserID), zap.Error(err))
				continue
			}
		}

		// Снимаем блокировку после завершения обработки заказа (processing = FALSE)
		err = s.storage.UnlockOrder(order.OrderNumber)
		if err != nil {
			s.logger.Error("Failed to unlock order", zap.String("order", order.OrderNumber), zap.Error(err))
			continue
		}
	}
}

func (s *OrderService) fetchOrderAccrual(orderNumber string) (float64, string, error) {
	url := s.accrualURL + "/api/orders/" + orderNumber

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		s.logger.Warn("Too many requests to accrual system", zap.String("order", orderNumber))
		return 0, domain.OrderStatusProcessing, nil
	}

	if resp.StatusCode == http.StatusNoContent {
		// Заказ не найден в системе начисления
		return 0, domain.OrderStatusInvalid, nil
	}

	if resp.StatusCode != http.StatusOK {
		return 0, "", gofermartErrors.ErrAccrualSystemUnavailable
	}

	var result struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, "", err
	}

	var status string
	switch result.Status {
	case "REGISTERED":
		status = domain.OrderStatusProcessing
	case "PROCESSING":
		status = domain.OrderStatusProcessing
	case "INVALID":
		status = domain.OrderStatusInvalid
	case "PROCESSED":
		status = domain.OrderStatusProcessed
	default:
		status = domain.OrderStatusProcessing
	}

	return result.Accrual, status, nil
}

func (s *OrderService) UpdateUserBalance(userID int, amount float64) error {
	err := s.storage.UpdateUserBalance(userID, amount)
	if err != nil {
		s.logger.Error("Failed to update user balance", zap.Int("userID", userID), zap.Error(err))
		return err
	}
	return nil
}
