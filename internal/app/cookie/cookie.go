package cookie

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/knstch/gophermart/internal/app/errorLogger"
)

func buildJWTString(login string, password string) (string, error) {
	const secretKey = "aboba"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":    login,
		"password": password,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		errorLogger.ErrorLogger("Error signing token: ", err)
		return "", err
	}

	return tokenString, nil
}

func SetAuth(res http.ResponseWriter, login string, password string) error {
	jwt, err := buildJWTString(login, password)
	if err != nil {
		errorLogger.ErrorLogger("Error making cookie: ", err)
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
