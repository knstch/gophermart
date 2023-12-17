// Package router provides a requests router.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/middleware/compressor"
	cookielogin "github.com/knstch/gophermart/internal/app/middleware/cookieLogin"
	statuslogger "github.com/knstch/gophermart/internal/app/middleware/statusLogger"
)

// func RequestsRouter(h *handler.Handler) chi.Router {
// 	router := chi.NewRouter()
// 	router.Use(
// 		statuslogger.WithLogger,
// 		compressor.WithCompressor,
// 	)
// 	router.Route("/", func(router chi.Router) {
// 		router.Route("/api", func(router chi.Router) {
// 			router.Route("/user", func(router chi.Router) {
// 				router.Post("/register", h.SignUp)
// 				router.Post("/login", h.Auth)
// 				router.Group(func(router chi.Router) {
// 					router.Use(cookielogin.WithCookieLogin)
// 					router.Get("/withdrawals", h.GetSpendOrderBonuses)
// 					router.Route("/orders", func(router chi.Router) {
// 						router.Post("/", h.UploadOrder)
// 						router.Get("/", h.GetOrders)
// 					})
// 					router.Route("/balance", func(router chi.Router) {
// 						router.Get("/", h.Balance)
// 						router.Post("/withdraw", h.WithdrawBonuses)
// 					})
// 				})
// 			})
// 		})
// 	})
// 	return router
// }

func RequestsRouter(h *handler.Handler) *gin.Engine {
    router := gin.Default()

    // Middleware
    router.Use(statuslogger.WithLogger())
    router.Use(compressor.WithCompressor())

    api := router.Group("/api")
    {
        user := api.Group("/user")
        {
            user.POST("/register", h.SignUp)
            user.POST("/login", h.Auth)

            user.Use(cookielogin.WithCookieLogin())
            {
                user.GET("/withdrawals", h.GetSpendOrderBonuses)

                orders := user.Group("/orders")
                {
                    orders.POST("/", h.UploadOrder)
                    orders.GET("/", h.GetOrders)
                }

                balance := user.Group("/balance")
                {
                    balance.GET("/", h.Balance)
                    balance.POST("/withdraw", h.WithdrawBonuses)
                }
            }
        }
    }

    return router
}