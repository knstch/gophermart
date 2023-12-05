package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgerrcode"
	"github.com/knstch/gophermart/internal/app/cookie"
	"github.com/knstch/gophermart/internal/app/errorLogger"
)

func NewHandler(s Storage) *Handler {
	return &Handler{s: s}
}

func (h *Handler) SignUp(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		errorLogger.ErrorLogger("Error during opening body: ", err)
	}
	var userData credentials

	err = json.Unmarshal(body, &userData)
	if err != nil {
		errorLogger.ErrorLogger("Error unmarshaling data: ", err)
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
		errorLogger.ErrorLogger("Can't set cookie: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Successfully registred"))
}
