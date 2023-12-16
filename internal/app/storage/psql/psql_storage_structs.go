package psql

import (
	"database/sql"
)

// A struct designed to insert login and password data to users table
type User struct {
	Login     string  `bun:"login"`
	Password  string  `bun:"password"`
	Balance   float32 `bun:"balance"`
	Withdrawn float32 `bun:"withdrawn"`
}

// A struct designed to insert login and order number to orders table
type Order struct {
	Login            string  `bun:"login"`
	Order            string  `bun:"order" json:"number"`
	Status           string  `bun:"status" json:"status"`
	UploadedAt       string  `bun:"uploaded_at" json:"uploaded_at"`
	BonusesWithdrawn *float32 `bun:"bonuses_withdrawn"`
	Accrual          *float32 `bun:"accrual" json:"accrual"`
}

type jsonOrder struct {
	Order            string  `json:"order"`
	Time             string  `json:"processed_at"`
	BonusesWithdrawn float32 `json:"sum"`
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
