// Package psql provides methods to interact with a Posgres database.
package psql

import (
	"context"
	"time"

	"github.com/knstch/gophermart/internal/app/common"
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

	bonusesWithdrawn := float32(0)

	userOrder := &common.Order{
		Login:            login,
		Order:            orderNum,
		UploadedAt:       now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: &bonusesWithdrawn,
	}

	isValid := validitycheck.LuhnAlgorithm(orderNum)
	if !isValid {
		return validitycheck.ErrWrongOrderNum
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	var checkOrder common.Order

	err := db.NewSelect().
		Model(&checkOrder).
		Where(`"order" = ?`, orderNum).
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
		return ErrAlreadyLoadedOrder
	} else if checkOrder.Login == login && checkOrder.Order == orderNum {
		return ErrYouAlreadyLoadedOrder
	}

	return nil
}

// GetOrders accepts context, login and returns an error and all user's orders
// ordered from old to new ones in json format.
func (storage *PsqURLlStorage) GetOrders(ctx context.Context, login string) ([]common.Order, error) {
	var allOrders []common.Order

	order := new(common.Order)

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
		var orderRow common.Order
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Status, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn, &orderRow.Accrual)
		if err != nil {
			logger.ErrorLogger("Error scanning data: ", err)
			return nil, err
		}
		allOrders = append(allOrders, common.Order{
			Order:      orderRow.Order,
			UploadedAt: orderRow.UploadedAt,
			Status:     orderRow.Status,
			Accrual:    orderRow.Accrual,
		})
	}

	return allOrders, nil
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
		return ErrNotEnoughBalance
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	checkOrder := new(common.Order)

	now := time.Now()

	userOrder := &common.Order{
		Login:            login,
		Order:            orderNum,
		UploadedAt:       now.Format(time.RFC3339),
		Status:           "NEW",
		BonusesWithdrawn: &spendBonuses,
	}

	err := db.NewSelect().
		Model(checkOrder).
		Where(`"order" = ?`, orderNum).
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
		return ErrAlreadyLoadedOrder
	} else if checkOrder.Login == login && checkOrder.Order == orderNum {
		return ErrYouAlreadyLoadedOrder
	}

	_, err = db.NewUpdate().
		TableExpr("users").
		Set("balance = ?", bonusesAvailable-spendBonuses).
		Set("withdrawn = withdrawn + ?", spendBonuses).
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
func (storage *PsqURLlStorage) GetOrdersWithBonuses(ctx context.Context, login string) ([]common.OrdersWithSpentBonuses, error) {
	var allOrders []common.OrdersWithSpentBonuses

	order := new(common.Order)

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
		var orderRow common.Order
		err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Status, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn, &orderRow.Accrual)
		if err != nil {
			logger.ErrorLogger("Error scanning data: ", err)
			return nil, err
		}

		allOrders = append(allOrders, common.OrdersWithSpentBonuses{
			Order:            orderRow.Order,
			Time:             orderRow.UploadedAt,
			BonusesWithdrawn: *orderRow.BonusesWithdrawn,
		})
	}

	if noRows {
		return nil, ErrNoRows
	}
	return allOrders, nil
}

// A function that looking for orders where bonuses were not
// received, sends them to the accrual system, updates status and tops up balance.
func (storage *PsqURLlStorage) Sync() {
	ticker := time.NewTicker(time.Second * 1)

	ctx := context.Background()

	for range ticker.C {
		var allUnfinishedOrders []common.Order

		order := new(common.Order)

		db := bun.NewDB(storage.db, pgdialect.New())

		rows, err := db.NewSelect().
			Model(order).
			Where("status != ? AND status != ?", "PROCESSED", "INVALID").
			Rows(ctx)
		rows.Err()
		if err != nil {
			logger.ErrorLogger("Error getting data: ", err)
		}

		for rows.Next() {
			var orderRow common.Order
			err := rows.Scan(&orderRow.Login, &orderRow.Order, &orderRow.Status, &orderRow.UploadedAt, &orderRow.BonusesWithdrawn, &orderRow.Accrual)
			if err != nil {
				logger.ErrorLogger("Error scanning data: ", err)
			}
			allUnfinishedOrders = append(allUnfinishedOrders, common.Order{
				Order:      orderRow.Order,
				UploadedAt: orderRow.UploadedAt,
				Status:     orderRow.Status,
				Accrual:    orderRow.Accrual,
				Login:      orderRow.Login,
			})
		}
		rows.Close()

		for _, unfinishedOrder := range allUnfinishedOrders {
			finishedOrder := common.GetStatusFromAccrual(unfinishedOrder)
			storage.UpdateStatus(ctx, finishedOrder, unfinishedOrder.Login)
		}
	}
}

// This function works with 2 tables: orders and users. As we get a status update from the accrual system,
// we make an update in the DB.
func (storage *PsqURLlStorage) UpdateStatus(ctx context.Context, orderFromAccural common.OrderUpdateFromAccural, login string) error {

	orderModel := &common.Order{}
	userModel := &User{}

	db := bun.NewDB(storage.db, pgdialect.New())

	_, err := db.NewUpdate().
		Model(orderModel).
		Set("status = ?, accrual = ?", orderFromAccural.Status, orderFromAccural.Accrual).
		Where(`"order" = ?`, orderFromAccural.Order).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error making an update request in order table", err)
		return err
	}

	_, err = db.NewUpdate().
		Model(userModel).
		Set("balance = balance + ?", orderFromAccural.Accrual).
		Where(`login = ?`, login).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error making an update request in user table", err)
		return err
	}
	return nil
}
