package balance

import "net/http"

type IndexGetHandler struct{}

// BalanceGetResponse описывает структуру ответа метода GET /api/user/balance
type BalanceGetResponse struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

func NewIndexGetHandler() *IndexGetHandler {
	return &IndexGetHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для получения текущего баланса счёта баллов лояльности пользователя.
func (h *IndexGetHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	// 200 — успешная обработка запроса.
	// 401 — пользователь не авторизован.
	// 500 — внутренняя ошибка сервера.
}
