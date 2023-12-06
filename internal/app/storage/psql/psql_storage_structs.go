package psql

import "errors"

type User struct {
	Login    string `bun:"login"`
	Password string `bun:"password"`
}

type Order struct {
	Login string `bun:"login"`
	Order int    `bun:"order_number"`
}

type Users struct {
	Login    string `bun:"type:varchar(255),unique"`
	Password string `bun:"type:varchar(255)"`
}

type Orders struct {
	Login       string `bun:"type:varchar(255)"`
	OrderNumber string `bun:"type:integer,unique"`
}

var ErrAlreadyLoadedOrder = errors.New("Order is loaded by another user")
