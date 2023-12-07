package psql

import (
	"context"
	"database/sql"
	"time"

	"github.com/knstch/gophermart/internal/app/logger"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// A functing receiving database params and initializing tables
// in the database. The function returns error in a case of any issues
// and nil if everything is good.
func InitDB(dbParams *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db := bun.NewDB(dbParams, pgdialect.New())

	_, err := db.NewCreateTable().Model((*Users)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error initing table Users: ", err)
		return err
	}

	_, err = db.NewCreateTable().Model((*Orders)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error initing table Orders: ", err)
		return err
	}

	logger.InfoLogger("Tables inited")

	return nil
}
