package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/knstch/gophermart/internal/app/middleware/compressor"
	"github.com/knstch/gophermart/internal/app/middleware/statusLogger"
	"github.com/knstch/gophermart/internal/app/storage/psql"
)

func RequestsRouter(h *handler.Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(statusLogger.RequestsLogger)
	router.Use(compressor.GzipMiddleware)
	router.Post("/api/user/register", h.SignUp)
	router.Post("/api/user/login", h.Auth)
	router.Post("/api/user/orders", h.UploadOrder)
	router.Get("/api/user/orders", h.GetOrders)
	router.Get("/api/user/balance", h.Balance)
	router.Post("/api/user/balance/withdraw", h.WithdrawBonuses)
	router.Get("/api/user/withdrawals", h.GetSpendOrderBonuses)
	return router
}

const psqlStorage = "host=localhost user=postgres password=Xer@0101 dbname=gophermart sslmode=disable"

func main() {
	db, err := sql.Open("pgx", psqlStorage)
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
		Addr:    ":8080",
		Handler: RequestsRouter(h),
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
