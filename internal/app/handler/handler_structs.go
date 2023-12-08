package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

// An interface responsible for operations with a database.
type Storage interface {
	Register(ctx context.Context, email string, password string) error
	CheckCredentials(ctx context.Context, login string, password string) error
	InsertOrder(ctx context.Context, login string, order string) error
	GetOrders(ctx context.Context, login string) ([]byte, error)
	GetBalance(ctx context.Context, login string) (int, int, error)
	SpendBonuses(ctx context.Context, login string, orderNum string, spendBonuses int) error
	GetOrdersWithBonuses(ctx context.Context, login string) ([]byte, error)
}

// A struct implementing Storage interface.
type Handler struct {
	s Storage
}

func NewHandler(s Storage) *Handler {
	return &Handler{s: s}
}

// A struct used to get and store data from json requests.
type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// A struct used to put data to json response
type balanceInfo struct {
	Balance   int `json:"balance"`
	Withdrawn int `json:"withdrawn"`
}

// A variable made to check an error type using errors.As
var pgErr *pgconn.PgError

type getSpendBonusRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}
