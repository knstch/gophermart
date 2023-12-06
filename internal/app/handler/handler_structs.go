package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

type Storage interface {
	Register(ctx context.Context, email string, password string) error
	CheckCredentials(ctx context.Context, login string, password string) error
	InsertOrder(ctx context.Context, login string, order int) error
}

type Handler struct {
	s Storage
}

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

var pgErr *pgconn.PgError
