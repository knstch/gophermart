// Package router provides a requests router.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/knstch/gophermart/internal/app/handler"

	// "github.com/knstch/gophermart/internal/app/middleware/compressor"
	cookielogin "github.com/knstch/gophermart/internal/app/middleware/cookieLogin"
	statuslogger "github.com/knstch/gophermart/internal/app/middleware/statusLogger"
)

func RequestsRouter(h *handler.Handler) *gin.Engine {
	router := gin.Default()

	router.Use(statuslogger.WithLogger())
	// router.Use(compressor.WithCompressor())

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
