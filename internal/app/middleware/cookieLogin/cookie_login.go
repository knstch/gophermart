package cookielogin

import (
	"context"
	"net/http"

	"github.com/knstch/gophermart/internal/app/cookie"
	gophermarterrors "github.com/knstch/gophermart/internal/app/gophermartErrors"
	"github.com/knstch/gophermart/internal/app/logger"
)

// A struct used to provide a value by a key in context.
type contextLogin struct {
	login string
}

// A type used to make a key in context.
type contextKey string

// A const containing a key called "login".
const LoginKey = contextKey("login")

// A builder function that returns a contextLogin struct with provided login as a param.
func newLogin(login string) *contextLogin {
	return &contextLogin{login: login}
}

// A middleware function checking if a user is logged in using cookie.
// If a URL path is not "/api/user/register" or "/api/user/login" and
// a user has a valid login cookie,
// it serves an https requests and inserts login inside of a context.
// Otherwise, it doesn't allow to go forward and returns 401 status code if
// a user is not authenticated or 500 if there is an Internal Server Error.
func WithCookieLogin(h http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		userLogin, err := cookie.GetCookie(req)
		if err == gophermarterrors.ErrAuth {
			logger.ErrorLogger("Error getting cookie", err)
			res.WriteHeader(401)
			res.Write([]byte("You are not authenticated"))
			return
		} else if err != nil {
			logger.ErrorLogger("Error reading cookie", err)
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte("Internal Server Error"))
			return
		}
		ctx := context.WithValue(req.Context(), LoginKey, newLogin(userLogin).login)
		req = req.WithContext(ctx)

		h.ServeHTTP(res, req)
	}
	return http.HandlerFunc(fn)
}
