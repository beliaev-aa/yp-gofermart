package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"beliaev-aa/yp-gofermart/internal/gofermart/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"go.uber.org/zap"
)

type OrdersGetHandler struct {
	logger            *zap.Logger
	orderService      *services.OrderService
	usernameExtractor utils.UsernameExtractor
}

func NewOrdersGetHandler(orderService *services.OrderService, usernameExtractor utils.UsernameExtractor, logger *zap.Logger) *OrdersGetHandler {
	return &OrdersGetHandler{
		logger:            logger,
		orderService:      orderService,
		usernameExtractor: usernameExtractor,
	}
}

type OrderResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float64 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func (o OrderResponse) MarshalJSON1() ([]byte, error) {
	var accrualField string
	if o.Accrual != 0 {
		accrualField = fmt.Sprintf(`"accrual": %.2f,`, o.Accrual)
	}

	jsonString := fmt.Sprintf(`{"number": "%s", "status": "%s", %s "uploaded_at": "%s"}`,
		o.Number,
		o.Status,
		accrualField,
		o.UploadedAt,
	)

	return []byte(jsonString), nil
}

// ServeHTTP обрабатывает HTTP-запросы для получения списка загруженных пользователем номеров заказов.
func (h *OrdersGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	login, err := h.usernameExtractor.ExtractUsernameFromContext(r, h.logger)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем список заказов пользователя
	orders, err := h.orderService.GetOrders(login)
	if err != nil {
		h.logger.Error("Failed to get orders", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Сортируем заказы по времени загрузки от старых к новым
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].UploadedAt.Before(orders[j].UploadedAt)
	})

	// Формируем ответ
	var response []OrderResponse
	for _, order := range orders {
		item := OrderResponse{
			Number:     order.OrderNumber,
			Status:     order.OrderStatus,
			UploadedAt: order.UploadedAt.Format(time.RFC3339),
		}
		if order.OrderStatus == domain.OrderStatusProcessed {
			item.Accrual = order.Accrual
		}
		response = append(response, item)
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
