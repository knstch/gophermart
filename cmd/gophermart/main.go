package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/knstch/gophermart/internal/app/errorLogger"
	"github.com/knstch/gophermart/internal/app/handler"
	"github.com/knstch/gophermart/internal/app/middleware/compressor"
	"github.com/knstch/gophermart/internal/app/middleware/logger"
	"github.com/knstch/gophermart/internal/app/storage/psql"
)

func RequestsRouter(h *handler.Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestsLogger)
	router.Use(compressor.GzipMiddleware)
	router.Post("/api/user/register", h.SignUp)
	return router
}

const psqlStorage = "host=localhost user=postgres password=Xer@0101 dbname=gophermart sslmode=disable"

func main() {
	db, err := sql.Open("pgx", psqlStorage)
	if err != nil {
		errorLogger.ErrorLogger("Can't open connection: ", err)
	}
	err = psql.InitDB(db)
	if err != nil {
		errorLogger.ErrorLogger("Can't init DB: ", err)
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
			errorLogger.ErrorLogger("Shutdown error: ", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		errorLogger.ErrorLogger("Server error: ", err)
	}
	<-idleConnsClosed
}
