package user

import "net/http"

type OrdersPostHandler struct{}

func NewOrdersPostHandler() *OrdersPostHandler {
	return &OrdersPostHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для загрузки пользователем номера заказа для расчёта.
func (h *OrdersPostHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	// 200 — номер заказа уже был загружен этим пользователем;
	// 202 — новый номер заказа принят в обработку;
	// 400 — неверный формат запроса;
	// 401 — пользователь не аутентифицирован;
	// 409 — номер заказа уже был загружен другим пользователем;
	// 422 — неверный формат номера заказа;
	// 500 — внутренняя ошибка сервера.
}
