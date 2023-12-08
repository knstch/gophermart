package psql

import (
	"context"
	"encoding/json"
	"time"

	gophermarterrors "github.com/knstch/gophermart/internal/app/gophermartErrors"
	"github.com/knstch/gophermart/internal/app/logger"
	validitycheck "github.com/knstch/gophermart/internal/app/validityCheck"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Register is used to add users to the database.
// It accepts login and password, inserts them to the database,
// and returns an error.
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

// CheckCredentials accepts login and password, then check if
// there is a match in the database. If nothing was found,
// it returns error.
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

// InsertOrder is used to insert information about an order to
// the database. It accepts login and order number and returns error.
// Before insering data, it checks the order number using Luhn algorithm,
// if the number is wrong, it returns a custom error.
func (storage *PsqURLlStorage) InsertOrder(ctx context.Context, login string, orderNum string) error {
	now := time.Now()

	userOrder := &Order{
		Login:        login,
		Order:        orderNum,
		Time:         now.Format(time.RFC3339),
		Status:       "NEW",
		SpentBonuses: 0,
	}

	isValid := validitycheck.LuhnAlgorithm(orderNum)
	if !isValid {
		return gophermarterrors.ErrWrongOrderNum
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
		return gophermarterrors.ErrAlreadyLoadedOrder
	} else if checkOrder.Login == login && checkOrder.Order == orderNum {
		return gophermarterrors.ErrYouAlreadyLoadedOrder
	}
	return nil
}

// GetOrders accepts login and returns an error and all user's orders
// ordered from old to new ones in json format.
func (storage *PsqURLlStorage) GetOrders(ctx context.Context, login string) ([]byte, error) {
	var allOrders []jsonOrder

	order := new(Order)

	db := bun.NewDB(storage.db, pgdialect.New())

	rows, err := db.NewSelect().
		Model(order).
		Where("login = ?", login).
		Order("uploaded_at ASC").
		Rows(ctx)
	rows.Err()
	if err != nil {
		logger.ErrorLogger("Error getting data: ", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var orderRow Order
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Time, &orderRow.Status, &orderRow.SpentBonuses)
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

// GetBalance accepts login and returns bonuses balance, withdraw
// amount, and error.
func (storage *PsqURLlStorage) GetBalance(ctx context.Context, login string) (int, int, error) {
	var user User

	db := bun.NewDB(storage.db, pgdialect.New())

	err := db.NewSelect().
		Model(&user).
		Where("login = ?", login).
		Scan(ctx)
	if err != nil {
		logger.ErrorLogger("Error finding user's balance: ", err)
		return 0, 0, err
	}

	return user.Balance, user.Withdrawn, nil
}

func (storage *PsqURLlStorage) SpendBonuses(ctx context.Context, login string, orderNum string, spendBonuses int) error {
	bonusesAvailable, _, nil := storage.GetBalance(ctx, login)
	if bonusesAvailable < spendBonuses {
		return gophermarterrors.ErrNotEnoughBalance
	}

	err := storage.InsertOrder(ctx, login, orderNum)
	if err != nil {
		return err
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	_, err = db.NewUpdate().
		TableExpr("orders").
		Set("bonuses_withdrawn = ?", spendBonuses).
		Where("order_number = ?", orderNum).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error speding bonuses: ", err)
		return err
	}

	_, err = db.NewUpdate().
		TableExpr("users").
		Set("balance = ?", bonusesAvailable-spendBonuses).
		Where("login = ?", login).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error withdrawning bonuses from the account: ", err)
		return err
	}
	return nil
}

func (storage *PsqURLlStorage) GetOrdersWithBonuses(ctx context.Context, login string) ([]byte, error) {
	var allOrders []jsonOrder

	order := new(Order)

	db := bun.NewDB(storage.db, pgdialect.New())

	rows, err := db.NewSelect().
		Model(order).
		Where("login = ? and bonuses_withdrawn != 0", login).
		Order("uploaded_at ASC").
		Rows(ctx)
	rows.Err()
	if err != nil {
		logger.ErrorLogger("Error getting data: ", err)
		return nil, err
	}
	defer rows.Close()

	noRows := true
	for rows.Next() {
		noRows = false
		var orderRow Order
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Time, &orderRow.Status, &orderRow.SpentBonuses)
		if err != nil {
			logger.ErrorLogger("Error scanning data: ", err)
			return nil, err
		}
		allOrders = append(allOrders, jsonOrder{
			Order:        orderRow.Order,
			Time:         orderRow.Time,
			SpentBonuses: orderRow.SpentBonuses,
		})
	}

	if noRows {
		return nil, gophermarterrors.ErrNoRows
	}

	jsonAllOrders, err := json.Marshal(allOrders)
	if err != nil {
		logger.ErrorLogger("Error marshaling orders: ", err)
	}

	return jsonAllOrders, nil
}
