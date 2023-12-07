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
	Login  string `bun:"login"`
	Order  string `bun:"order_number"`
	Time   string `bun:"uploaded_at"`
	Status string `bun:"status"`
}

// A struct used to convert data to JSON
type jsonOrder struct {
	Order  string `json:"number"`
	Time   string `json:"uploaded_at"`
	Status string `json:"status"`
}

// A struct designed to initialize users table in the database
type Users struct {
	Login     string `bun:"type:varchar(255),unique"`
	Password  string `bun:"type:varchar(255)"`
	Balance   int    `type:"integer"`
	Withdrawn int    `type:"integer"`
}

// A struct designed to initialize orders table in the database
type Orders struct {
	Login       string `bun:"type:varchar(255)"`
	OrderNumber string `bun:"type:varchar(255),unique"`
	Status      string `bun:"type:varchar(255)"`
	UploadedAt  string `bun:"type:timestamp"`
}

// A struct used to set database connection and
// bind database interaction methods
type PsqURLlStorage struct {
	db *sql.DB
}
