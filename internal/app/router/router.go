package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/middleware/compressor"
	cookielogin "github.com/knstch/gophermart/internal/app/middleware/cookieLogin"
	statuslogger "github.com/knstch/gophermart/internal/app/middleware/statusLogger"
)

func RequestsRouter(h *handler.Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(
		statuslogger.WithLogger,
		compressor.WithCompressor,
	)
	router.Route("/", func(router chi.Router) {
		router.Route("/api", func(router chi.Router) {
			router.Route("/user", func(router chi.Router) {
				router.Post("/register", h.SignUp)
				router.Post("/login", h.Auth)
				router.Group(func(router chi.Router) {
					router.Use(cookielogin.WithCookieLogin)
					router.Get("/withdrawals", h.GetSpendOrderBonuses)
					router.Route("/orders", func(router chi.Router) {
						router.Post("/", h.UploadOrder)
						router.Get("/", h.GetOrders)
					})
					router.Route("/balance", func(router chi.Router) {
						router.Get("/", h.Balance)
						router.Post("/withdraw", h.WithdrawBonuses)
					})
				})
			})
		})
	})
	return router
}
