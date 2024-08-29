package user

import (
	"net/http"
)

type OrdersGetHandler struct{}

// OrdersGetResponse описывает структуру ответа метода
type OrdersGetResponse struct {
	Number     string  `json:"number"`
	Status     string  `json:"status"`
	Accrual    float32 `json:"accrual,omitempty"`
	UploadedAt string  `json:"uploaded_at"`
}

func NewOrdersGetHandler() *OrdersGetHandler {
	return &OrdersGetHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для получения списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях.
func (h *OrdersGetHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	// TODO: Вынести в domain
	// NEW — заказ загружен в систему, но не попал в обработку;
	// PROCESSING — вознаграждение за заказ рассчитывается;
	// INVALID — система расчёта вознаграждений отказала в расчёте;
	// PROCESSED — данные по заказу проверены и информация о расчёте успешно получена.

	// 200 — успешная обработка запроса.
	// 204 — нет данных для ответа.
	// 401 — пользователь не авторизован.
	// 500 — внутренняя ошибка сервера.
}
