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

// OrdersPostHandler - представляет HTTP-обработчик для загрузки номера заказа пользователем.
type OrdersPostHandler struct {
	logger            *zap.Logger
	orderService      services.OrderServiceInterface
	usernameExtractor utils.UsernameExtractor
}

// NewOrdersPostHandler - создает новый экземпляр OrdersPostHandler с указанными зависимостями.
func NewOrdersPostHandler(orderService services.OrderServiceInterface, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *OrdersPostHandler {
	return &OrdersPostHandler{
		logger:            logger,
		orderService:      orderService,
		usernameExtractor: usernameExtractor,
	}
}

// ServeHTTP - обрабатывает HTTP-запросы POST для загрузки номера заказа.
func (h *OrdersPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, err := h.usernameExtractor.ExtractUsernameFromContext(r, h.logger)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		h.logger.Warn("Invalid request body", zap.Error(err))
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body", zap.Error(err))
		}
	}()

	orderNumber := strings.TrimSpace(string(body))

	if !utils.IsValidMoon(orderNumber) {
		h.logger.Warn("Invalid order number format", zap.String("orderNumber", orderNumber))
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

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
