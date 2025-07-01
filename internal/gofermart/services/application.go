package services

type AppServices struct {
	AuthService  *AuthService
	OrderService OrderServiceInterface
	UserService  *UserService
}
