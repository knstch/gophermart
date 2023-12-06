package psql

import (
	"context"
	"database/sql"

	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type PsqURLlStorage struct {
	db *sql.DB
}

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

	err := db.NewSelect().Model(&user).Where("login = ? and password = ?", login, password).Scan(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (storage *PsqURLlStorage) InsertOrder(ctx context.Context, login string, order int) error {
	userOrder := &Order{
		Login: login,
		Order: order,
	}

	db := bun.NewDB(storage.db, pgdialect.New())

	checkOrder := new(Order)

	err := db.NewSelect().
		Model(checkOrder).
		Where("order_number = ?", order).
		Scan(ctx)
	if err != nil {
		_, err = db.NewInsert().
			Model(userOrder).
			Exec(ctx)

		if err != nil {
			logger.ErrorLogger("Error writing data: ", err)
			return err
		}
		return nil
	}
	if checkOrder.Login != login && checkOrder.Order == order {
		return ErrAlreadyLoadedOrder
	}
	return nil
}
