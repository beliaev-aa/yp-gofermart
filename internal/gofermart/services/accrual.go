package services

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// AccrualService - представляет интерфейс для работы с внешней системой начислений
type AccrualService interface {
	GetOrderAccrual(ctx context.Context, orderNumber string) (float64, string, error)
}

// RealAccrualService - реализация интерфейса AccrualService
type RealAccrualService struct {
	BaseURL string
	logger  *zap.Logger
	limiter *rate.Limiter
	mu      sync.Mutex
}

// NewAccrualService - конструктор для RealAccrualService
func NewAccrualService(BaseURL string, logger *zap.Logger) AccrualService {
	return &RealAccrualService{
		BaseURL: BaseURL,
		logger:  logger,
		limiter: rate.NewLimiter(rate.Inf, 1), // Изначально без ограничения
	}
}

// GetOrderAccrual - основная функция для получения информации о заказе и обработки ответа
func (s *RealAccrualService) GetOrderAccrual(ctx context.Context, orderNumber string) (float64, string, error) {
	// Ожидаем разрешения от лимитера
	if err := s.limiter.Wait(ctx); err != nil {
		s.logger.Error("Limiter error", zap.Error(err))
		return 0, "", fmt.Errorf("limiter error: %w", err)
	}

	// Формируем запрос
	url := s.BaseURL + "/api/orders/" + orderNumber
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		s.logger.Error("Failed to create new request", zap.Error(err))
		return 0, "", fmt.Errorf("failed to create new request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Request to accrual system failed", zap.Error(err))
		return 0, "", fmt.Errorf("request to accrual system failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Error("Failed to close response body", zap.Error(closeErr))
		}
	}()

	// Обновляем настройки лимитера на основе заголовков ответа
	s.updateRateLimiter(resp.Header)

	// Обрабатываем коды ответа HTTP
	switch resp.StatusCode {
	case http.StatusTooManyRequests:
		s.logger.Warn("Too many requests to accrual system", zap.String("order", orderNumber))

		// Получаем заголовок Retry-After
		retryAfter := resp.Header.Get("Retry-After")

		// Разбираем заголовок Retry-After
		var waitDuration time.Duration
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				waitDuration = time.Duration(seconds) * time.Second
			} else if t, err := http.ParseTime(retryAfter); err == nil {
				waitDuration = time.Until(t)
			}
		} else {
			// Если заголовок отсутствует, ждем по умолчанию 1 минуту
			waitDuration = time.Minute
		}

		// Обновляем лимитер, блокируя запросы на указанный период
		s.mu.Lock()
		s.limiter.SetLimit(0)
		s.mu.Unlock()

		// Планируем снятие блокировки после истечения времени ожидания
		go func() {
			time.Sleep(waitDuration)
			s.mu.Lock()
			s.limiter.SetLimit(rate.Inf)
			s.mu.Unlock()
		}()

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
		return 0, "", fmt.Errorf("failed to decode JSON response: %w", err)
	}

	switch result.Status {
	case domain.OrderStatusRegistered,
		domain.OrderStatusProcessing,
		domain.OrderStatusInvalid,
		domain.OrderStatusProcessed:
		// Возвращаем статус как есть
	default:
		s.logger.Error("Received unknown order status from the accrual system", zap.String("status", result.Status))
		return 0, "", errors.New("received unknown order status from the accrual system")
	}

	return result.Accrual, result.Status, nil
}

// updateRateLimiter - обновляет настройки лимитера на основе заголовков ответа
func (s *RealAccrualService) updateRateLimiter(headers http.Header) {
	rateLimit := headers.Get("X-RateLimit-Limit")
	rateRemaining := headers.Get("X-RateLimit-Remaining")
	rateReset := headers.Get("X-RateLimit-Reset")

	if rateLimit != "" && rateRemaining != "" && rateReset != "" {
		limit, err1 := strconv.ParseFloat(rateLimit, 64)
		remaining, err2 := strconv.ParseFloat(rateRemaining, 64)
		resetUnix, err3 := strconv.ParseInt(rateReset, 10, 64)

		if err1 == nil && err2 == nil && err3 == nil {
			now := time.Now()
			resetTime := time.Unix(resetUnix, 0)
			duration := resetTime.Sub(now)

			if duration <= 0 {
				duration = time.Second
			}

			requestsPerSecond := remaining / duration.Seconds()
			if requestsPerSecond <= 0 {
				requestsPerSecond = 1 / duration.Seconds()
			}

			s.mu.Lock()
			s.limiter.SetLimit(rate.Limit(requestsPerSecond))
			s.limiter.SetBurst(int(limit))
			s.mu.Unlock()
		}
	}
}
