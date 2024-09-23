package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

// AccrualService - представляет интерфейс для работы с внешней системой начислений
type AccrualService interface {
	GetOrderAccrual(orderNumber string) (float64, string, error)
}

// RealAccrualService - реализация интерфейса AccrualService
type RealAccrualService struct {
	BaseURL string
	logger  *zap.Logger
}

func NewAccrualService(BaseURL string, logger *zap.Logger) AccrualService {
	return &RealAccrualService{
		BaseURL: BaseURL,
		logger:  logger,
	}
}

// GetOrderAccrual - делает запрос во внешнюю систему для получения информации о заказе
func (s *RealAccrualService) GetOrderAccrual(orderNumber string) (float64, string, error) {
	url := s.BaseURL + "/api/orders/" + orderNumber

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, "", err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.logger.Error("Failed to close response", zap.Error(err))
			return
		}
	}(resp.Body)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			s.logger.Error("Failed to close body response", zap.Error(err))
		}
	}(resp.Body)

	var status string
	switch result.Status {
	case "REGISTERED", "PROCESSING":
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
