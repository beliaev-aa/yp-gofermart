package httpserver

import (
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server/handlers/api/user"
	"beliaev-aa/yp-gofermart/internal/gofermart/http-server/handlers/api/user/balance"
	"beliaev-aa/yp-gofermart/internal/gofermart/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"
)

// RegisterRoutes регистрирует роуты приложения
func RegisterRoutes(r *chi.Mux, appServices *services.AppServices, logger *zap.Logger) {
	JWTAuth := appServices.AuthService.GetTokenAuth()
	compressMiddleware := middleware.Compress(5, "gzip", "deflate")

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", user.NewRegisterPostHandler(appServices.AuthService, logger).ServeHTTP)
			r.Post("/login", user.NewLoginPostHandler(appServices.AuthService, logger).ServeHTTP)

			r.Group(func(r chi.Router) {
				r.Use(jwtauth.Verifier(JWTAuth))
				r.Use(jwtauth.Authenticator(JWTAuth))

				r.Post("/orders", user.NewOrdersPostHandler(appServices.OrderService, logger).ServeHTTP)
				r.With(compressMiddleware).Get("/orders", user.NewOrdersGetHandler(appServices.OrderService, logger).ServeHTTP)
				r.Route("/balance", func(r chi.Router) {
					r.With(compressMiddleware).Get("/", balance.NewIndexGetHandler(appServices.UserService, logger).ServeHTTP)
					r.Post("/withdraw", balance.NewWithdrawPostHandler(appServices.UserService, logger).ServeHTTP)
				})
				r.With(compressMiddleware).Get("/withdrawals", user.NewWithdrawalsGetHandler(appServices.UserService, logger).ServeHTTP)
			})
		})
	})
}
