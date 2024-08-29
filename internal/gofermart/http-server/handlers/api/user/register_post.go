package user

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"encoding/json"
	"net/http"
)

type RegisterPostHandler struct{}

func NewRegisterPostHandler() *RegisterPostHandler {
	return &RegisterPostHandler{}
}

// ServeHTTP обрабатывает HTTP-запросы для регистрации пользователя.
func (h *RegisterPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request domain.AuthenticationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверка занятости логина, если занят вернуть 409 ошибку

	// 200 — пользователь успешно зарегистрирован и аутентифицирован;
	// 400 — неверный формат запроса;
	// 409 — логин уже занят;
	// 500 — внутренняя ошибка сервера.
}
