// Package router provides a requests router.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/knstch/gophermart/internal/app/handler"

	"github.com/gin-contrib/gzip"
	_ "github.com/knstch/gophermart/docs"
	cookielogin "github.com/knstch/gophermart/internal/app/middleware/cookieLogin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router serving requests
func RequestsRouter(h *handler.Handler) *gin.Engine {
	router := gin.Default()

	router.Use(gzip.Gzip(gzip.BestCompression))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := router.Group("/api")
	{
		user := api.Group("/user")
		{
			user.POST("/register", h.SignUp)
			user.POST("/login", h.Auth)

			user.Use(cookielogin.WithCookieLogin())
			{
				user.GET("/withdrawals", h.GetOrderWithSpentBonuses)

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
