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

// @Summary SignUp
// @Tags Auth
// @Description API for user registration and setting an auth cookie
// @Accept json
// @Produce json
// @Param userData body credentials true "Login and password"
// @Success 200 {object} Message "Successfully registered"
// @Failure 400 {object} ErrorMessage "Wrong request"
// @Failure 409 {object} ErrorMessage "Login is already taken"
// @Failure 500 {object} ErrorMessage "Internal Server Error"
// @Router /user/register [post]
func (h *Handler) SignUp(ctx *gin.Context) {
	var userData Credentials

	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, newErrorMessage("Wrong request"))
		return
	}

	if userData.Login == "" || userData.Password == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, newErrorMessage("Wrong request"))
		return
	}

	err = h.s.Register(ctx, userData.Login, userData.Password)
	switch {
	case errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code):
		ctx.AbortWithStatusJSON(http.StatusConflict, newErrorMessage("Login is already taken"))
		return
	case err != nil:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}
	err = cookie.SetAuth(ctx.Writer, userData.Login)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}
	ctx.JSON(http.StatusOK, newMessage("Successfully registered"))
}

// @Summary Auth
// @Tags Auth
// @Description API for user authentication and setting an auth cookie
// @Accept json
// @Produce json
// @Param userData body credentials true "Login and password"
// @Success 200 {object} Message "Successfully signed in"
// @Failure 400 {object} ErrorMessage "Wrong request"
// @Failure 401 {object} ErrorMessage "Wrong email or password"
// @Failure 500 {object} ErrorMessage "Internal Server Error"
// @Router /user/login [post]
func (h *Handler) Auth(ctx *gin.Context) {
	var userData Credentials

	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&userData)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, newErrorMessage("Wrong request"))
		return
	}

	err = h.s.CheckCredentials(ctx, userData.Login, userData.Password)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, newErrorMessage("Wrong email or password"))
		return
	}

	err = cookie.SetAuth(ctx.Writer, userData.Login)
	if err != nil {
		logger.ErrorLogger("Can't set cookie: ", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}

	ctx.JSON(http.StatusOK, newMessage("Successfully signed in"))
}

// @Summary Upload order
// @Tags Order
// @Description Uploads an order to the server
// @Accept plain
// @Produce json
// @Param orderNum body string true "Order number"
// @Success 200 {object} Message "Order was successfully loaded before"
// @Success 202 {object} Message "Order was successfully accepted"
// @Failure 409 {object} ErrorMessage "Order is already loaded"
// @Failure 422 {object} ErrorMessage "Wrong order number"
// @Failure 500 {object} ErrorMessage "Internal Server Error"
// @Router /user/orders [post]
func (h *Handler) UploadOrder(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}

	orderNum := string(body)

	login := ctx.Value("login").(string)

	err = h.s.InsertOrder(ctx, login, orderNum)
	switch {
	case errors.Is(err, psql.ErrAlreadyLoadedOrder):
		ctx.AbortWithStatusJSON(http.StatusConflict, newErrorMessage("Order is already loaded by another user"))
		return
	case errors.Is(err, psql.ErrYouAlreadyLoadedOrder):
		ctx.AbortWithStatusJSON(http.StatusOK, newMessage("Order is already loaded"))
		return
	case errors.Is(err, validitycheck.ErrWrongOrderNum):
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, newErrorMessage("Wrong order number"))
		return
	default:
		ctx.JSON(http.StatusAccepted, newMessage("Successfully loaded order"))
	}
}

// @Summary Get user's orders
// @Description Retrieves the orders associated with the user
// @Tags Order
// @Produce json
// @Success 200 {array} common.Order "A list of user's orders"
// @Failure 204 {object} Message "A user has no orders"
// @Failure 500 {object} ErrorMessage "Internal server error"
// @Router /user/orders [get]
func (h *Handler) GetOrders(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	orders, err := h.s.GetOrders(ctx, login)
	if err != nil {
		logger.ErrorLogger("Error getting orders", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}

	if len(orders) == 0 {
		ctx.AbortWithStatusJSON(http.StatusNoContent, newMessage("You have no orders"))
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// @Summary Get user's balance
// @Description Retrieves the user's balance and withdrawn amount
// @Tags Balance
// @Produce json
// @Success 200 {object} balanceInfo "User's balance"
// @Failure 500 {object} ErrorMessage "Internal Server Error"
// @Router /user/balance [get]
func (h *Handler) Balance(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	balance, withdrawn, err := h.s.GetBalance(ctx, login)
	if err != nil {
		logger.ErrorLogger("Error getting balance", err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}

	userBalance := balanceInfo{
		Balance:   balance,
		Withdrawn: withdrawn,
	}

	ctx.JSON(http.StatusOK, userBalance)
}

// @Summary Withdraw user's bonuses
// @Description Allows users to spend their bonuses
// @Tags Balance
// @Accept json
// @Produce json
// @Param orderData body getSpendBonusRequest true "Order number and withdraw amount"
// @Success 200 {object} Message "Bonuses successfully spent"
// @Failure 400 {object} ErrorMessage "Wrong request"
// @Failure 402 {object} ErrorMessage "Not enough balance"
// @Failure 422 {object} ErrorMessage "Wrong order number"
// @Failure 409 {object} ErrorMessage "Order is already loaded"
// @Failure 500 {object} ErrorMessage "Internal Server Error"
// @Router /user/balance/withdraw [post]
func (h *Handler) WithdrawBonuses(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	var spendRequest getSpendBonusRequest

	if err := ctx.ShouldBindJSON(&spendRequest); err != nil {
		logger.ErrorLogger("Wrong request: ", err)
		ctx.JSON(http.StatusBadRequest, newErrorMessage("Wrong request"))
		return
	}

	err := h.s.SpendBonuses(ctx, login, spendRequest.Order, spendRequest.Sum)
	switch {
	case errors.Is(err, psql.ErrNotEnoughBalance):
		ctx.AbortWithStatusJSON(http.StatusPaymentRequired, newErrorMessage("Not enough balance"))
		return
	case errors.Is(err, validitycheck.ErrWrongOrderNum):
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, newErrorMessage("Wrong order number"))
		return
	case errors.Is(err, psql.ErrAlreadyLoadedOrder) || errors.Is(err, psql.ErrYouAlreadyLoadedOrder):
		ctx.AbortWithStatusJSON(http.StatusConflict, newErrorMessage("Order is already loaded"))
		return
	case err != nil:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}
	ctx.JSON(http.StatusOK, newMessage("Bonuses successfully spent"))
}

// @Summary Get orders with spent bonuses
// @Description Retrieves the orders with bonuses spent by the user
// @Tags Order
// @Produce json
// @Success 200 {object} common.OrdersWithSpentBonuses "A list of orders with spent bonuses"
// @Failure 204 {object} Message "You have not spent any bonuses"
// @Failure 500 {object} ErrorMessage "Internal Server Error"
// @Router /user/withdrawals [get]
func (h *Handler) GetOrderWithSpentBonuses(ctx *gin.Context) {
	login := ctx.Value("login").(string)

	ordersWithBonuses, err := h.s.GetOrdersWithBonuses(ctx, login)
	switch {
	case errors.Is(err, psql.ErrNoRows):
		ctx.AbortWithStatusJSON(http.StatusNoContent, newMessage("You have not spent any bonuses"))
		return
	case err != nil:
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, newErrorMessage("Internal Server Error"))
		return
	}

	ctx.JSON(http.StatusOK, ordersWithBonuses)
}
