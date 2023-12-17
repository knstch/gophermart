package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/knstch/gophermart/internal/app/common"
)

// An interface responsible for operations with a database.
type Storage interface {
	Register(ctx context.Context, email string, password string) error
	CheckCredentials(ctx context.Context, login string, password string) error
	InsertOrder(ctx context.Context, login string, order string) error
	GetOrders(ctx context.Context, login string) ([]common.Order, error)
	GetBalance(ctx context.Context, login string) (float32, float32, error)
	SpendBonuses(ctx context.Context, login string, orderNum string, spendBonuses float32) error
	GetOrdersWithBonuses(ctx context.Context, login string) ([]common.OrdersWithSpentBonuses, error)
}

// A struct implementing Storage interface.
type Handler struct {
	s Storage
}

// A builder function returning a Handler struct with Storage interface.
func NewHandler(s Storage) *Handler {
	return &Handler{s: s}
}

// A struct used to get and store data from a json requests.
type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// A struct used to put data to a json response
type balanceInfo struct {
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

// A variable made to check an error type using errors.As
var pgErr *pgconn.PgError

// A struct used to parse a json request to withdraw bonuses making an order.
type getSpendBonusRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}
