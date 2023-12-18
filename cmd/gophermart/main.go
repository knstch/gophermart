package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/router"
	"github.com/knstch/gophermart/internal/app/storage/psql"
)

// @title Gophermart API
// @version 1.0
// @description API server for users to sign up, sign it, upload orders, get and spend bonuses, and check balance

// @host localhost:8080
// @BasePath /api
// @securitydefinitions.apikey ApiKeyAuth
// @in cookie
// @name Auth

func main() {
	config.ParseConfig()
	db, err := sql.Open("pgx", config.ReadyConfig.Database)
	if err != nil {
		logger.ErrorLogger("Can't open connection: ", err)
	}
	err = psql.InitDB(db)
	if err != nil {
		logger.ErrorLogger("Can't init DB: ", err)
	}

	storage := psql.NewPsqlStorage(db)

	h := handler.NewHandler(storage)

	srv := http.Server{
		Addr:    config.ReadyConfig.ServerAddr,
		Handler: router.RequestsRouter(h),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.ErrorLogger("Shutdown error: ", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.ErrorLogger("Server error: ", err)
	}
	<-idleConnsClosed
}
