package cookielogin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/knstch/gophermart/internal/app/cookie"
	gophermarterrors "github.com/knstch/gophermart/internal/app/gophermartErrors"
	"github.com/knstch/gophermart/internal/app/logger"
)

type contextLogin struct {
	login string
}

type contextKey string

const LoginKey = contextKey("login")

func newLogin(login string) *contextLogin {
	return &contextLogin{login: login}
}

func WithCookieLogin(h http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/user/register" && req.URL.Path != "/api/user/login" {
			userLogin, err := cookie.GetCookie(req)
			if err == gophermarterrors.ErrAuth {
				logger.ErrorLogger("Error getting cookie", err)
				res.WriteHeader(401)
				res.Write([]byte("You are not authenticated"))
				fmt.Println("Aboba")
				return
			} else if err != nil {
				logger.ErrorLogger("Error reading cookie", err)
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("Internal Server Error"))
				return
			}
			ctx := context.WithValue(req.Context(), LoginKey, newLogin(userLogin).login)
			req = req.WithContext(ctx)
		}

		h.ServeHTTP(res, req)
	}
	return http.HandlerFunc(fn)
}
