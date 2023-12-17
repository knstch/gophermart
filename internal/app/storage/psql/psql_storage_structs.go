package psql

import (
	"database/sql"
	"errors"
)

// A struct designed to insert login and password data to users table
type User struct {
	Login     string  `bun:"login"`
	Password  string  `bun:"password"`
	Balance   float32 `bun:"balance"`
	Withdrawn float32 `bun:"withdrawn"`
}

// A struct designed to initialize users table in the database
type Users struct {
	Login     string  `bun:"type:varchar(255),unique"`
	Password  string  `bun:"type:varchar(255)"`
	Balance   float32 `bun:"type:float"`
	Withdrawn float32 `bun:"type:float"`
}

// A struct designed to initialize orders table in the database
type Orders struct {
	Login            string  `bun:"type:varchar(255)"`
	Order            string  `bun:"type:varchar(255),unique"`
	Status           string  `bun:"type:varchar(255)"`
	UploadedAt       string  `bun:"type:timestamp"`
	BonusesWithdrawn float32 `bun:"type:float"`
	Accrual          float32 `bun:"type:float"`
}

// A struct used to set database connection and
// bind database interaction methods
type PsqURLlStorage struct {
	db *sql.DB
}

// A builder function used in main.go file made to initialize Postgres storage
// with its methods
func NewPsqlStorage(db *sql.DB) *PsqURLlStorage {
	return &PsqURLlStorage{db: db}
}

type OrderUpdateFromAccural struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

// An error indicating that an order is loaded by another user.
var ErrAlreadyLoadedOrder = errors.New("order is loaded by another user")

// An error indicating that an order is loaded by user.
var ErrYouAlreadyLoadedOrder = errors.New("order is loaded by you")

// An error indictating that a user has not enough balance.
var ErrNotEnoughBalance = errors.New("not enough balance")

// An error indiating that no rows were found.
var ErrNoRows = errors.New("no rows were found")
