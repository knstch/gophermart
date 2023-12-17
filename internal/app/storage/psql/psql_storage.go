// Package psql provides methods to interact with a Posgres database.
package psql

import (
	"context"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/common"
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

	go storage.GetStatusFromAccural(*userOrder)

	return nil
}

// This function is designed to work asynchronously, it accepts an order and
// creates 2 channels, orderJob working with Order type and result
// working with OrderUpdateFromAccural type. As we put an order to orderJob,
// we trigger a goroutine doing Get requests to the accrual system until we get
// "INVALID" or "PROCESSED" status. On each order status change we put an updated
// data of OrderUpdateFromAccural type to the result channel and trigger
// a function updating information in the DB.
func (storage *PsqURLlStorage) GetStatusFromAccural(order common.Order) {
	var wg sync.WaitGroup

	orderJob := make(chan common.Order)
	result := make(chan OrderUpdateFromAccural)

	defer close(result)

	wg.Add(1)
	go func(jobs <-chan common.Order, result chan<- OrderUpdateFromAccural) {

		defer wg.Done()

		client := resty.New().SetBaseURL(config.ReadyConfig.Accural)
		job := <-jobs
		lastResult := OrderUpdateFromAccural{}
		for {
			var orderUpdate OrderUpdateFromAccural

			resp, err := client.R().
				SetResult(&orderUpdate).
				Get("/api/orders/" + job.Order)
			if err != nil {
				logger.ErrorLogger("Got error trying to send a get request from worker: ", err)
				break
			}
			switch resp.StatusCode() {
			case 429:
				time.Sleep(3 * time.Second)
			case 204:
				time.Sleep(1 * time.Second)
			}

			if resp.StatusCode() == 500 {
				logger.ErrorLogger("Internal server error in accural system: ", err)
				break
			}
			if orderUpdate != lastResult {
				lastResult = orderUpdate
				result <- lastResult
			}
			if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
	}(orderJob, result)

	orderJob <- order
	defer close(orderJob)

	go func() {
		for orderToUpdate := range result {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := storage.UpdateStatus(ctx, orderToUpdate, order.Login)
			if err != nil {
				logger.ErrorLogger("Error updating status: ", err)
			}
			cancel()
		}
	}()

	wg.Wait()
}

// This function works with 2 tables: orders and users. As we get a status update from the accrual system,
// we make an update in the DB.
func (storage *PsqURLlStorage) UpdateStatus(ctx context.Context, order OrderUpdateFromAccural, login string) error {

	orderModel := new(common.Order)
	userModel := new(User)

	db := bun.NewDB(storage.db, pgdialect.New())

	_, err := db.NewUpdate().
		Model(orderModel).
		Set("status = ?, accrual = ?", order.Status, order.Accrual).
		Where(`"order" = ?`, order.Order).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error making an update request in order table", err)
		return err
	}

	_, err = db.NewUpdate().
		Model(userModel).
		Set("balance = balance + ?", order.Accrual).
		Where(`login = ?`, login).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error making an update request in user table", err)
		return err
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
