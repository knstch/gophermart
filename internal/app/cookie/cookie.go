package cookie

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/knstch/gophermart/cmd/config"
	gophermarterrors "github.com/knstch/gophermart/internal/app/gophermartErrors"
	"github.com/knstch/gophermart/internal/app/logger"
)

// A claim struct containing jwt.RegisteredClaims and Login
type Claims struct {
	jwt.RegisteredClaims
	Login string
}

// A function building a JWT token and retrning this token and error.
func buildJWTString(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
	})

	tokenString, err := token.SignedString([]byte(config.ReadyConfig.SecretKey))
	if err != nil {
		logger.ErrorLogger("Error signing token: ", err)
		return "", err
	}

	return tokenString, nil
}

// A functing setting an auth JWT token in cookies. It accepts http.ResponseWriter and login and returns an error.
func SetAuth(res http.ResponseWriter, login string) error {
	jwt, err := buildJWTString(login)
	if err != nil {
		logger.ErrorLogger("Error making cookie: ", err)
		return err
	}

	cookie := http.Cookie{
		Name:  "Auth",
		Value: jwt,
		Path:  "/",
	}
	http.SetCookie(res, &cookie)

	return nil
}

// A function used to get a user's login using a JWT. It accepts a JWT and returns a login and error.
func getLogin(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.ErrorLogger("unexpected signing method", nil)
			return nil, nil
		}
		return []byte(config.ReadyConfig.SecretKey), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		logger.ErrorLogger("Token is not valid", nil)
		return "", err
	}
	return claims.Login, nil
}

// A function used to get a cookie and return a login and error.
func GetCookie(req *http.Request) (string, error) {
	signedLogin, err := req.Cookie("Auth")
	if err != nil {
		logger.ErrorLogger("Error getting cookie", err)
		return "", gophermarterrors.ErrAuth
	}

	login, err := getLogin(signedLogin.Value)
	if err != nil {
		logger.ErrorLogger("Error reading cookie", err)
		return "", err
	}

	return login, nil
}
