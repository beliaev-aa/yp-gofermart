package domain

// AuthenticationRequest - описывает запрос на регистрацию и аутентификацию пользователя
type AuthenticationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
