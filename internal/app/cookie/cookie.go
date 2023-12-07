package cookie

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/customErrors"
)

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

func buildJWTString(login string, password string) (string, error) {
	const secretKey = "aboba"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		logger.ErrorLogger("Error signing token: ", err)
		return "", err
	}

	return tokenString, nil
}

func SetAuth(res http.ResponseWriter, login string, password string) error {
	jwt, err := buildJWTString(login, password)
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

func getLogin(tokenString string) (string, error) {
	const secretKey = "aboba"
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.ErrorLogger("unexpected signing method", nil)
			return nil, nil
		}
		return []byte(secretKey), nil
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

func GetCookie(req *http.Request) (string, error) {
	signedLogin, err := req.Cookie("Auth")
	if err != nil {
		logger.ErrorLogger("Error getting cookie", err)
		return "", customErrors.ErrAuth
	}

	login, err := getLogin(signedLogin.Value)
	if err != nil {
		logger.ErrorLogger("Error reading cookie", err)
		return "", err
	}

	return login, nil
}
