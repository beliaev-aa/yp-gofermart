package httpserver

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server/handlers/api/user"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server/handlers/api/user/balance"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes регистрирует роуты приложения
func RegisterRoutes(r *chi.Mux) {
	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", user.NewRegisterPostHandler().ServeHTTP)
			r.Post("/login", user.NewLoginPostHandler().ServeHTTP)
			r.Post("/orders", user.NewOrdersPostHandler().ServeHTTP)
			r.Get("/orders", user.NewOrdersGetHandler().ServeHTTP)
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balance.NewIndexGetHandler().ServeHTTP)
				r.Post("/withdraw", balance.NewWithdrawPostHandler().ServeHTTP)
			})
			r.Get("/withdrawals", user.NewWithdrawalsGetHandler().ServeHTTP)
		})
	})

}
