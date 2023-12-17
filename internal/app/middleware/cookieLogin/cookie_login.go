// Package cookielogin provides a middleware to parse a user login.
package cookielogin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knstch/gophermart/internal/app/cookie"
	"github.com/knstch/gophermart/internal/app/logger"
)

// A middleware function checking if a user is logged in using cookie.
// If a URL path is not "/api/user/register" or "/api/user/login" and
// a user has a valid login cookie,
// it serves an https requests and inserts login inside of a context.
// Otherwise, it doesn't allow to go forward and returns 401 status code if
// a user is not authenticated or 500 if there is an Internal Server Error.
func WithCookieLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userLogin, err := cookie.GetCookie(ctx.Request)
		if err == cookie.ErrAuth {
			logger.ErrorLogger("Error getting cookie", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "You are not authenticated"})
		} else if err != nil {
			logger.ErrorLogger("Error reading cookie", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}
		ctx.Set("login", userLogin)
		ctx.Next()
	}
}
