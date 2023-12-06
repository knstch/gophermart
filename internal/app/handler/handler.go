package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/jackc/pgerrcode"
	"github.com/knstch/gophermart/internal/app/cookie"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/storage/psql"
)

func NewHandler(s Storage) *Handler {
	return &Handler{s: s}
}

func (h *Handler) SignUp(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}
	var userData credentials

	err = json.Unmarshal(body, &userData)
	if err != nil {
		logger.ErrorLogger("Error unmarshaling data: ", err)
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Wrong request"))
		return
	}

	err = h.s.Register(req.Context(), userData.Login, userData.Password)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		res.WriteHeader(409)
		res.Write([]byte("Login is already taken"))
		return
	} else if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}
	err = cookie.SetAuth(res, userData.Login, userData.Password)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Successfully registred"))
}

func (h *Handler) Auth(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}

	var userData credentials

	err = json.Unmarshal(body, &userData)
	if err != nil {
		logger.ErrorLogger("Error unmarshaling data: ", err)
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Wrong request"))
		return
	}

	err = h.s.CheckCredentials(req.Context(), userData.Login, userData.Password)

	if err != nil {
		logger.ErrorLogger("Wrong email or password: ", err)
		res.WriteHeader(401)
		res.Write([]byte("Wrong email or password"))
		return
	}

	err = cookie.SetAuth(res, userData.Login, userData.Password)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Successfully signed in"))
}

func (h *Handler) UploadOrder(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	orderNum, err := strconv.Atoi(string(body))
	if err != nil {
		logger.ErrorLogger("Error converting order to num: ", err)
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Wrong order number format"))
		return
	}

	signedLogin, err := req.Cookie("Auth")
	if err != nil {
		logger.ErrorLogger("Error getting cookie", err)
		res.WriteHeader(401)
		res.Write([]byte("You are not authenticated"))
		return
	}

	login, err := cookie.GetLogin(signedLogin.Value)
	if err != nil {
		logger.ErrorLogger("Error reading cookie", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	err = h.s.InsertOrder(req.Context(), login, orderNum)
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		res.WriteHeader(200)
		res.Write([]byte("Order is already loaded"))
		return
	} else if errors.Is(err, psql.ErrAlreadyLoadedOrder) {
		res.WriteHeader(409)
		res.Write([]byte("Order is already loaded by another user"))
		return
	}

	res.WriteHeader(202)
	res.Write([]byte("Successfully loaded ordred"))
}
