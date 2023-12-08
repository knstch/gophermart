package psql

import (
	"database/sql"
)

// A struct designed to insert login and password data to users table
type User struct {
	Login     string `bun:"login"`
	Password  string `bun:"password"`
	Balance   int    `bun:"balance"`
	Withdrawn int    `bun:"withdrawn"`
}

// A struct designed to insert login and order number to orders table
type Order struct {
	Login        string `bun:"login"`
	Order        string `bun:"order_number"`
	Time         string `bun:"uploaded_at"`
	Status       string `bun:"status"`
	SpentBonuses int    `bun:"bonuses_withdrawn"`
}

// A struct used to convert data to JSON
type jsonOrder struct {
	Order        string `json:"number"`
	Time         string `json:"uploaded_at"`
	Status       string `json:"status"`
	SpentBonuses int    `json:"sum"`
}

// A struct designed to initialize users table in the database
type Users struct {
	Login     string `bun:"type:varchar(255),unique"`
	Password  string `bun:"type:varchar(255)"`
	Balance   int    `bun:"type:float"`
	Withdrawn int    `bun:"type:integer"`
}

// A struct designed to initialize orders table in the database
type Orders struct {
	Login            string `bun:"type:varchar(255)"`
	OrderNumber      string `bun:"type:varchar(255),unique"`
	Status           string `bun:"type:varchar(255)"`
	UploadedAt       string `bun:"type:timestamp"`
	BonusesWithdrawn int    `bun:"type:integer"`
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
