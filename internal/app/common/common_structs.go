package common

// A struct designed to insert data to order table
type Order struct {
	Login            string   `bun:"login" json:"-"`
	Order            string   `bun:"order" json:"number"`
	Status           string   `bun:"status" json:"status"`
	UploadedAt       string   `bun:"uploaded_at" json:"uploaded_at"`
	BonusesWithdrawn *float32 `bun:"bonuses_withdrawn"`
	Accrual          *float32 `bun:"accrual" json:"accrual"`
}

// A struct designed to return data to a client about orders with withdrawn bonuses
type OrdersWithSpentBonuses struct {
	Order            string  `json:"order"`
	Time             string  `json:"processed_at"`
	BonusesWithdrawn float32 `json:"sum"`
}

// A struct designed to receive data from accrual system
type OrderUpdateFromAccural struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}
