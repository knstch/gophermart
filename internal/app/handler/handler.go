// Package handler is used to serve http requests.
package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/knstch/gophermart/internal/app/cookie"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/storage/psql"
	validitycheck "github.com/knstch/gophermart/internal/app/validityCheck"
)

// A handler used to sign up a user setting an auth cookie.
func (h *Handler) SignUp(ctx *gin.Context) {
	var userData credentials

	if err := ctx.ShouldBindJSON(&userData); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong request"})
	}

	err := h.s.Register(ctx, userData.Login, userData.Password)
	switch {
	case errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code):
		ctx.AbortWithStatusJSON(409, gin.H{"error": "Login is already taken"})
	case err != nil:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
	err = cookie.SetAuth(ctx.Writer, userData.Login)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully registered"})
}

// A handler used to authenticate a user setting an auth cookie.
func (h *Handler) Auth(ctx *gin.Context) {
	var userData credentials

	if err := ctx.ShouldBindJSON(&userData); err != nil {
		logger.ErrorLogger("Wrong request: ", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Wrong request"})
	}

	err := h.s.CheckCredentials(ctx, userData.Login, userData.Password)
	if err != nil {
		logger.ErrorLogger("Wrong email or password: ", err)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong email or password"})
	}

	err = cookie.SetAuth(ctx.Writer, userData.Login)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully signed in"})
}

// A handler used to upload a user's order.
func (h *Handler) UploadOrder(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

	orderNum := string(body)

	login := ctx.Value("login").(string)

	err = h.s.InsertOrder(ctx, login, orderNum)
	switch {
	case errors.Is(err, psql.ErrAlreadyLoadedOrder):
		ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Order is already loaded by another user"})
	case errors.Is(err, psql.ErrYouAlreadyLoadedOrder):
		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{"message": "Order is already loaded"})
	case errors.Is(err, validitycheck.ErrWrongOrderNum):
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "Wrong order number"})
	default:
		ctx.JSON(http.StatusAccepted, gin.H{"message": "Successfully loaded order"})
	}
}

// A handler used to get all user's orders.
func (h *Handler) GetOrders(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	orders, err := h.s.GetOrders(ctx, login)
	if err != nil {
		logger.ErrorLogger("Error getting orders", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

	if len(orders) == 4 {
		ctx.AbortWithStatusJSON(http.StatusNoContent, gin.H{"message": "You have no orders"})
	}

	ctx.JSON(http.StatusOK, orders)
}

// A handler used to check user's balance.
func (h *Handler) Balance(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	balance, withdrawn, err := h.s.GetBalance(ctx, login)
	if err != nil {
		logger.ErrorLogger("Error getting balance", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

	userBalance := balanceInfo{
		Balance:   balance,
		Withdrawn: withdrawn,
	}

	jsonUserBalance, err := json.Marshal(userBalance)
	if err != nil {
		logger.ErrorLogger("Error marshaling json", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

	ctx.JSON(http.StatusOK, jsonUserBalance)
}

// A handler allowing a user to make an order using bonuses.
func (h *Handler) WithdrawBonuses(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	var spendRequest getSpendBonusRequest

	if err := ctx.ShouldBindJSON(&spendRequest); err != nil {
		logger.ErrorLogger("Wrong request: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wrong request"})
		return
	}

	err := h.s.SpendBonuses(ctx, login, spendRequest.Order, spendRequest.Sum)
	switch {
	case errors.Is(err, psql.ErrNotEnoughBalance):
		ctx.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{"error": "Not enough balance"})
	case errors.Is(err, validitycheck.ErrWrongOrderNum):
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"error": "Wrong order number"})
	case errors.Is(err, psql.ErrAlreadyLoadedOrder) || errors.Is(err, psql.ErrYouAlreadyLoadedOrder):
		ctx.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Order is already loaded"})
	case err != nil:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
	ctx.JSON(http.StatusAccepted, gin.H{"message": "Bonuses successfully spent"})
}

// A handler used to get all user's orders with spent bonuses.
func (h *Handler) GetSpendOrderBonuses(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	ordersWithBonuses, err := h.s.GetOrdersWithBonuses(ctx, login)
	switch {
	case errors.Is(err, psql.ErrNoRows):
		ctx.AbortWithStatusJSON(http.StatusNoContent, gin.H{"message": "You have not spent any bonuses"})
	case err != nil:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
	}

	ctx.JSON(http.StatusOK, ordersWithBonuses)
}
