package psql

import (
	"context"
	"database/sql"
	"time"

	"github.com/knstch/gophermart/internal/app/errorLogger"
)

func InitDB(db *sql.DB) error {

	tx, err := db.Begin()
	if err != nil {
		errorLogger.ErrorLogger("Proccess a transaction: ", err)
		return err
	}
	initialization := `CREATE TABLE IF NOT EXISTS users(
		 login varchar(255) UNIQUE,
		 password varchar(255));`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, initialization)
	if err != nil {
		tx.Rollback()
		errorLogger.ErrorLogger("Can't exec: ", err)
		return err
	}

	errorLogger.InfoLogger("Tables inited")

	return tx.Commit()
}
