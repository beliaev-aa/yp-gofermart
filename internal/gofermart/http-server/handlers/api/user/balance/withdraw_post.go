package balance

import "net/http"

type WithdrawPostHandler struct{}

// WithdrawPostResponse описывает структуру ответа метода POST /api/user/balance/withdraw
type WithdrawPostResponse struct {
	Order int     `json:"order"`
	Sum   float32 `json:"sum"`
}

func NewWithdrawPostHandler() *WithdrawPostHandler {
	return &WithdrawPostHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для запроса на списание баллов с накопительного счёта в счёт оплаты нового заказа.
func (h *WithdrawPostHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	// 200 — успешная обработка запроса;
	// 401 — пользователь не авторизован;
	// 402 — на счету недостаточно средств;
	// 422 — неверный номер заказа;
	// 500 — внутренняя ошибка сервера.
}
