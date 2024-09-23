package user

import (
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type OrdersPostHandler struct {
	logger            *zap.Logger
	orderService      *services.OrderService
	usernameExtractor utils.UsernameExtractor
}

func NewOrdersPostHandler(orderService *services.OrderService, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *OrdersPostHandler {
	return &OrdersPostHandler{
		logger:            logger,
		orderService:      orderService,
		usernameExtractor: usernameExtractor,
	}
}

// ServeHTTP обрабатывает HTTP-запросы для загрузки пользователем номера заказа для расчёта.
func (h *OrdersPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, err := h.usernameExtractor.ExtractUsernameFromContext(r, h.logger)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		h.logger.Warn("Invalid request body", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			h.logger.Error("Failed to close response", zap.Error(err))
		}
	}(r.Body)

	orderNumber := strings.TrimSpace(string(body))

	// Проверяем формат номера заказа с помощью алгоритма Луна
	if !utils.IsValidMoon(orderNumber) {
		h.logger.Warn("Invalid order number format", zap.String("orderNumber", orderNumber))
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// Добавляем заказ
	err = h.orderService.AddOrder(login, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, gofermartErrors.ErrOrderAlreadyUploaded):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, gofermartErrors.ErrOrderUploadedByAnother):
			http.Error(w, "Order number already uploaded by another user", http.StatusConflict)
			return
		default:
			h.logger.Error("Failed to add order", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Успешно приняли заказ в обработку
	w.WriteHeader(http.StatusAccepted)
}
