package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/knstch/gophermart/internal/app/cookie"
	getbonuses "github.com/knstch/gophermart/internal/app/getBonuses"
	gophermarterrors "github.com/knstch/gophermart/internal/app/gophermartErrors"
	"github.com/knstch/gophermart/internal/app/logger"
	cookielogin "github.com/knstch/gophermart/internal/app/middleware/cookieLogin"
)

// A handler used to sign up a user setting an auth cookie.
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
	err = cookie.SetAuth(res, userData.Login)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Successfully registred"))
}

// A handler used to authenticate a user setting an auth cookie.
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

	err = cookie.SetAuth(res, userData.Login)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Successfully signed in"))
}

// A handler used to upload a user's order.
func (h *Handler) UploadOrder(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	orderNum := string(body)

	login := req.Context().Value(cookielogin.LoginKey).(string)

	err = h.s.InsertOrder(req.Context(), login, orderNum)
	if errors.Is(err, gophermarterrors.ErrAlreadyLoadedOrder) {
		res.WriteHeader(409)
		res.Write([]byte("Order is already loaded by another user"))
		return
	} else if errors.Is(err, gophermarterrors.ErrYouAlreadyLoadedOrder) {
		res.WriteHeader(200)
		res.Write([]byte("Order is already loaded"))
		return
	} else if errors.Is(err, gophermarterrors.ErrWrongOrderNum) {
		res.WriteHeader(422)
		res.Write([]byte("Wrong order number"))
		return
	}
	go getbonuses.GetStatusFromAccural(orderNum, login)
	time.Sleep(1 * time.Second)
	res.WriteHeader(202)
	res.Write([]byte("Successfully loaded ordred"))
}

// A handler used to get all user's orders.
func (h *Handler) GetOrders(res http.ResponseWriter, req *http.Request) {
	login := req.Context().Value(cookielogin.LoginKey).(string)

	res.Header().Set("Content-Type", "application/json")
	orders, err := h.s.GetOrders(req.Context(), login)
	if err != nil {
		logger.ErrorLogger("Error getting orders", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	if len(orders) == 4 {
		res.WriteHeader(204)
		res.Write([]byte("You have no orders"))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(orders))
}

// A handler used to check user's balance.
func (h *Handler) Balance(res http.ResponseWriter, req *http.Request) {
	login := req.Context().Value(cookielogin.LoginKey).(string)

	balance, withdrawn, err := h.s.GetBalance(req.Context(), login)
	if err != nil {
		logger.ErrorLogger("Error getting balance", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
	}

	userBalance := balanceInfo{
		Balance:   balance,
		Withdrawn: withdrawn,
	}

	jsonUserBalance, err := json.Marshal(userBalance)
	if err != nil {
		logger.ErrorLogger("Error marshaling json", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(jsonUserBalance))
}

// A handler allowing a user to make an order using bonuses.
func (h *Handler) WithdrawBonuses(res http.ResponseWriter, req *http.Request) {
	login := req.Context().Value(cookielogin.LoginKey).(string)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}

	var spendRequest getSpendBonusRequest

	err = json.Unmarshal(body, &spendRequest)
	if err != nil {
		logger.ErrorLogger("Error unmarshaling data: ", err)
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Wrong request"))
		return
	}

	err = h.s.SpendBonuses(req.Context(), login, spendRequest.Order, spendRequest.Sum)
	if errors.Is(err, gophermarterrors.ErrNotEnoughBalance) {
		res.WriteHeader(402)
		res.Write([]byte("Not enough balance"))
		return
	} else if errors.Is(err, gophermarterrors.ErrWrongOrderNum) {
		res.WriteHeader(422)
		res.Write([]byte("Wrong order number"))
		return
	} else if errors.Is(err, gophermarterrors.ErrAlreadyLoadedOrder) || errors.Is(err, gophermarterrors.ErrYouAlreadyLoadedOrder) {
		res.WriteHeader(409)
		res.Write([]byte("Order is already loaded"))
		return
	} else if err != nil {
		logger.ErrorLogger("Error spending bonuses", err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("Internal Server Error"))
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Bonuses successfully spent"))
}

// A handler used to get all user's orders with spent bonuses.
func (h *Handler) GetSpendOrderBonuses(res http.ResponseWriter, req *http.Request) {
	login := req.Context().Value(cookielogin.LoginKey).(string)

	ordersWithBonuses, err := h.s.GetOrdersWithBonuses(req.Context(), login)
	if errors.Is(err, gophermarterrors.ErrNoRows) {
		res.WriteHeader(204)
		res.Write([]byte("You have not spent any bonuses"))
		return
	} else if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte("err"))
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(ordersWithBonuses))
}
