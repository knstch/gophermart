package psql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	customErrors "github.com/knstch/gophermart/internal/app/customErrors"
	"github.com/knstch/gophermart/internal/app/logger"
	validitycheck "github.com/knstch/gophermart/internal/app/validityCheck"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// A builder function used in main.go file made to initialize Postgres storage
// with its methods
func NewPsqlStorage(db *sql.DB) *PsqURLlStorage {
	return &PsqURLlStorage{db: db}
}

func (storage *PsqURLlStorage) Register(ctx context.Context, login string, password string) error {
	credentials := &User{
		Login:    login,
		Password: password,
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	_, err := db.NewInsert().
		Model(credentials).
		Exec(ctx)

	if err != nil {
		logger.ErrorLogger("Error writing data: ", err)
		return err
	}

	return nil
}

func (storage *PsqURLlStorage) CheckCredentials(ctx context.Context, login string, password string) error {
	var user User

	db := bun.NewDB(storage.db, pgdialect.New())

	err := db.NewSelect().
		Model(&user).
		Where("login = ? and password = ?", login, password).
		Scan(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (storage *PsqURLlStorage) InsertOrder(ctx context.Context, login string, orderNum string) error {
	now := time.Now()

	userOrder := &Order{
		Login:  login,
		Order:  orderNum,
		Time:   now.Format(time.RFC3339),
		Status: "NEW",
	}

	isValid := validitycheck.LuhnAlgorithm(orderNum)
	if !isValid {
		return customErrors.ErrWrongOrderNum
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	checkOrder := new(Order)

	err := db.NewSelect().
		Model(checkOrder).
		Where("order_number = ?", orderNum).
		Scan(ctx)
	if err != nil {
		_, err := db.NewInsert().
			Model(userOrder).
			Exec(ctx)

		if err != nil {
			logger.ErrorLogger("Error writing data: ", err)
			return err
		}
		return nil
	}
	if checkOrder.Login != login && checkOrder.Order == orderNum {
		return customErrors.ErrAlreadyLoadedOrder
	} else if checkOrder.Login == login && checkOrder.Order == orderNum {
		return customErrors.ErrYouAlreadyLoadedOrder
	}
	return nil
}

func (storage *PsqURLlStorage) GetOrders(ctx context.Context, login string) ([]byte, error) {
	var allOrders []jsonOrder

	order := new(Order)

	db := bun.NewDB(storage.db, pgdialect.New())

	rows, err := db.NewSelect().
		Model(order).
		Where("login = ?", login).
		Rows(ctx)
	rows.Err()
	if err != nil {
		logger.ErrorLogger("Error getting data: ", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var orderRow Order
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Time, &orderRow.Status)
		if err != nil {
			logger.ErrorLogger("Error scanning data: ", err)
			return nil, err
		}
		allOrders = append(allOrders, jsonOrder{
			Order:  orderRow.Order,
			Time:   orderRow.Time,
			Status: orderRow.Status,
		})
	}

	jsonAllOrders, err := json.Marshal(allOrders)
	if err != nil {
		logger.ErrorLogger("Error marshaling orders: ", err)
	}

	return jsonAllOrders, nil
}

func (storage *PsqURLlStorage) GetBalance(ctx context.Context, login string) (int, int, error) {
	var user User

	db := bun.NewDB(storage.db, pgdialect.New())

	err := db.NewSelect().
		Model(&user).
		Where("login = ?", login).
		Scan(ctx)
	if err != nil {
		return 0, 0, nil
	}

	return user.Balance, user.Withdrawn, nil
}
