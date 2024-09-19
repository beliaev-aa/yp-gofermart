package user

import (
	gofermartErrors "beliaev-aa/yp-gofermart/internal/gofermart/errors"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"errors"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type OrdersPostHandler struct {
	orderService *services.OrderService
	logger       *zap.Logger
}

func NewOrdersPostHandler(orderService *services.OrderService, logger *zap.Logger) *OrdersPostHandler {
	return &OrdersPostHandler{
		orderService: orderService,
		logger:       logger,
	}
}

// ServeHTTP обрабатывает HTTP-запросы для загрузки пользователем номера заказа для расчёта.
func (h *OrdersPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о пользователе из JWT-токена
	_, claims, _ := jwtauth.FromContext(r.Context())
	login, ok := claims["username"].(string)
	if !ok {
		h.logger.Warn("Unauthorized access attempt")
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
	defer r.Body.Close()

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
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Успешно приняли заказ в обработку
	w.WriteHeader(http.StatusAccepted)
}
