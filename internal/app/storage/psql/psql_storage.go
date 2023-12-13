package psql

import (
	"context"
	"encoding/json"
	"time"

	getbonuses "github.com/knstch/gophermart/internal/app/getBonuses"
	gophermarterrors "github.com/knstch/gophermart/internal/app/gophermartErrors"
	"github.com/knstch/gophermart/internal/app/logger"
	validitycheck "github.com/knstch/gophermart/internal/app/validityCheck"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Register is used to add users to the database.
// It accepts context, login and password, inserts them to the database,
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

// CheckCredentials accepts context, login and password, then check if
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
// the database. It accepts context, login and order number and returns error.
// Before insering data, it checks the order number using Luhn algorithm,
// if the number is wrong, it returns a custom error.
func (storage *PsqURLlStorage) InsertOrder(ctx context.Context, login string, orderNum string) error {
	now := time.Now()

	userOrder := &Order{
		Login:            login,
		Order:            orderNum,
		Time:             now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: 0,
		Accural:          0,
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
	if checkOrder.Login != login && checkOrder.Order == orderNum {
		return gophermarterrors.ErrAlreadyLoadedOrder
	} else if checkOrder.Login == login && checkOrder.Order == orderNum {
		return gophermarterrors.ErrYouAlreadyLoadedOrder
	}
	if err != nil {
		_, err := db.NewInsert().
			Model(userOrder).
			Exec(ctx)

		if err != nil {
			logger.ErrorLogger("Error writing data: ", err)
			return err
		}

		err = getbonuses.GetStatusFromAccural(userOrder.Order)
		if err != nil {
			logger.ErrorLogger("Error getting bonuses: ", err)
		}
		return nil
	}

	return nil
}

// GetOrders accepts context, login and returns an error and all user's orders
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
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Time, &orderRow.Status, &orderRow.BonusesWithdrawn, &orderRow.Accural)
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

// GetBalance accepts context and login, and returns bonuses balance, withdraw
// amount, and error.
func (storage *PsqURLlStorage) GetBalance(ctx context.Context, login string) (float32, float32, error) {
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

// SpendBonuses accepts context, login, order number, and amount of bonuses to spend.
// It allows to spend user's bonuses on an order.
// This function returns error in an error case or nil if everything is good.
func (storage *PsqURLlStorage) SpendBonuses(ctx context.Context, login string, orderNum string, spendBonuses float32) error {
	bonusesAvailable, _, nil := storage.GetBalance(ctx, login)
	if bonusesAvailable < spendBonuses {
		return gophermarterrors.ErrNotEnoughBalance
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	checkOrder := new(Order)

	now := time.Now()

	userOrder := &Order{
		Login:            login,
		Order:            orderNum,
		Time:             now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: spendBonuses,
		Accural:          0,
	}

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
	}
	if checkOrder.Login != login && checkOrder.Order == orderNum {
		return gophermarterrors.ErrAlreadyLoadedOrder
	} else if checkOrder.Login == login && checkOrder.Order == orderNum {
		return gophermarterrors.ErrYouAlreadyLoadedOrder
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

// This function accepts context and login, and returns an error and json response with orders where a user
// spent bonuses.
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
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Time, &orderRow.Status, &orderRow.BonusesWithdrawn)
		if err != nil {
			logger.ErrorLogger("Error scanning data: ", err)
			return nil, err
		}
		allOrders = append(allOrders, jsonOrder{
			Order:        orderRow.Order,
			Time:         orderRow.Time,
			SpentBonuses: orderRow.BonusesWithdrawn,
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
