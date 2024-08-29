package httpserver

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/domain"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server/handlers/api/user"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server/handlers/api/user/balance"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RegisterRoutes регистрирует роуты приложения
func RegisterRoutes(r *chi.Mux, cfg *domain.Config, logger *zap.Logger) {
	authService := services.NewAuthService([]byte(cfg.JWTSecret), logger)
	r.Use(middleware.Compress(5, "gzip", "deflate"))

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", user.NewRegisterPostHandler(authService, logger).ServeHTTP)
			r.Post("/login", user.NewLoginPostHandler(authService, logger).ServeHTTP)
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
