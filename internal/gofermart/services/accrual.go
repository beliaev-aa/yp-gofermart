package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
)

// AccrualService - представляет интерфейс для работы с внешней системой начислений
type AccrualService interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (float64, string, error)
}

// RealAccrualService - реализация интерфейса AccrualService
type RealAccrualService struct {
	BaseURL string
	logger  *zap.Logger
}

// NewAccrualService - конструктор для RealAccrualService
func NewAccrualService(BaseURL string, logger *zap.Logger) AccrualService {
	return &RealAccrualService{
		BaseURL: BaseURL,
		logger:  logger,
	}
}

// GetOrderAccrual - основная функция для получения информации о заказе и обработки ответа
func (s *RealAccrualService) GetOrderAccrual(ctx context.Context, orderNumber string) (float64, string, error) {
	url := s.BaseURL + "/api/orders/" + orderNumber
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create new request", zap.Error(err))
		return 0, "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Request to accrual system failed", zap.Error(err))
		return 0, "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	// Обрабатываем коды ответа HTTP
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		s.logger.Warn("Too many requests to accrual system", zap.String("order", orderNumber))
		return 0, domain.OrderStatusProcessing, nil
	case http.StatusNoContent:
		return 0, domain.OrderStatusInvalid, nil
	case http.StatusOK:
		// Продолжаем обработку
	default:
		s.logger.Error("Accrual system returned an error", zap.Int("status", resp.StatusCode))
		return 0, "", gofermartErrors.ErrAccrualSystemUnavailable
	}

	// Декодируем JSON-ответ
	var result struct {
		Order   string  `json:"order"`
		Status  string  `json:"status"`
		Accrual float64 `json:"accrual,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.logger.Error("Failed to decode JSON response", zap.Error(err))
		return 0, "", err
	}

	switch result.Status {
	case "REGISTERED", "PROCESSING", "INVALID", "PROCESSED":
		// Возвращаем статус как есть
	default:
		s.logger.Error("Received unknown order status from the accrual system", zap.String("status", result.Status))
		return 0, "", errors.New("received unknown order status from the accrual system")
	}

	return result.Accrual, result.Status, nil
}
