package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"encoding/json"
	"net/http"
)

type LoginPostHandler struct{}

func NewLoginPostHandler() *LoginPostHandler {
	return &LoginPostHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для аутентификации пользователя.
func (h *LoginPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request domain.AuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 200 — пользователь успешно аутентифицирован;
	// 400 — неверный формат запроса;
	// 401 — неверная пара логин/пароль;
	// 500 — внутренняя ошибка сервера.
}
