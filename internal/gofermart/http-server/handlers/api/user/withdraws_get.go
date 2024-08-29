package user

import "net/http"

type WithdrawalsGetHandler struct{}

// WithdrawGetResponse описывает структуру ответа метода GET /api/user/withdrawals
type WithdrawGetResponse struct {
	Order       int     `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func NewWithdrawalsGetHandler() *WithdrawalsGetHandler {
	return &WithdrawalsGetHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для получения информации о выводе средств с накопительного счёта пользователем.
func (h *WithdrawalsGetHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	// 200 — успешная обработка запроса.
	// 204 — нет ни одного списания.
	// 401 — пользователь не авторизован.
	// 500 — внутренняя ошибка сервера.
}
