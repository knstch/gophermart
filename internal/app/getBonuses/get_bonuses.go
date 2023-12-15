package getbonuses

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type Storage interface {
	UpdateStatus(ctx context.Context, order OrderUpdateFromAccural, login string) error
}

type StatusUpdater struct {
	s Storage
}

type OrderUpdateFromAccural struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

type OrderToAccuralSys struct {
	Order string
}

func NewOrderToAccuralSys(order string) OrderToAccuralSys {
	return OrderToAccuralSys{
		Order: order,
	}
}

func NewStatusUpdater(s Storage) *StatusUpdater {
	return &StatusUpdater{s: s}
}

type PsqURLlStorage struct {
	db *sql.DB
}

func NewPsqlStorage(db *sql.DB) *PsqURLlStorage {
	return &PsqURLlStorage{db: db}
}

// Semaphore структура семафора
type Semaphore struct {
	semaCh chan struct{}
}

// NewSemaphore создает семафор с буферизованным каналом емкостью maxReq
func NewSemaphore(maxReq int) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

// когда горутина запускается, отправляем пустую структуру в канал semaCh
func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// когда горутина завершается, из канала semaCh убирается пустая структура
func (s *Semaphore) Release() {
	<-s.semaCh
}

type Order struct {
	Login            string  `bun:"login" json:"-"`
	Order            string  `bun:"order" json:"order"`
	Time             string  `bun:"uploaded_at" json:"uploaded_at"`
	Status           string  `bun:"status" json:"status"`
	BonusesWithdrawn float32 `bun:"bonuses_withdrawn" json:"sum"`
	Accural          float32 `bun:"accural" json:"-"`
}

func (storage *PsqURLlStorage) UpdateStatus(ctx context.Context, order OrderUpdateFromAccural, login string) error {
	fmt.Println("Acquaired works??: ", order.Accrual)
	ord := Order{
		Login:   login,
		Order:   order.Order,
		Status:  order.Status,
		Accural: order.Accrual,
	}
	db := bun.NewDB(storage.db, pgdialect.New())
	_, err := db.NewUpdate().
		Model(&ord).
		Set(`status = ?`, ord.Status).
		Set(`accural = ?`, ord.Accural).
		Where(`order = ?`, ord.Order).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error withdrawning bonuses from the account: ", err)
		return err
	}

	var orderPosted Order
	_, err = db.NewSelect().Model(&orderPosted).Where(`"order" = ?`, order.Order).Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error checking order: ", err)
		return err
	}
	fmt.Println("Order after post! ", orderPosted.Accural)

	_, err = db.NewUpdate().
		TableExpr("users").
		Set(`balance = ?`, order.Accrual).
		Where(`"login" = ?`, login).
		Exec(ctx)
	if err != nil {
		logger.ErrorLogger("Error topping up the balance: ", err)
		return err
	}
	return nil
}

func GetStatusFromAccural(order string, login string) {
	fmt.Println("GetStatusFromAccural works")
	db, err := sql.Open("pgx", config.ReadyConfig.Database)
	if err != nil {
		logger.ErrorLogger("Error setting the connection with the database: ", err)
	}
	storage := NewPsqlStorage(db)
	updater := NewStatusUpdater(storage)

	var wg sync.WaitGroup

	semaphore := NewSemaphore(5)

	sendOrderToJobs := NewOrderToAccuralSys(order)
	OrderJob := make(chan OrderToAccuralSys)
	result := make(chan OrderUpdateFromAccural)

	defer close(result)

	for idx := 0; idx < 5; idx++ {
		wg.Add(1)
		go func(jobs <-chan OrderToAccuralSys, result chan<- OrderUpdateFromAccural) {

			semaphore.Acquire()
			defer wg.Done()
			defer semaphore.Release()

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
				fmt.Println("Resp status code: ", resp.StatusCode())
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
				fmt.Println("Order status: ", orderUpdate.Status)
				if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
					break
				}
				time.Sleep(250 * time.Millisecond)
			}

		}(OrderJob, result)
	}

	OrderJob <- sendOrderToJobs
	defer close(OrderJob)

	go func() {
		for orderToUpdate := range result {
			fmt.Println("Triggered result chan", orderToUpdate.Accrual)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			updater.s.UpdateStatus(ctx, orderToUpdate, login)
			cancel()
		}
	}()

	wg.Wait()
}

// func GetStatusFromAccural(order string) error {
// 	db, err := sql.Open("pgx", config.ReadyConfig.Database)
// 	if err != nil {
// 		logger.ErrorLogger("Error setting the connection with the database: ", err)
// 	}
// 	storage := NewPsqlStorage(db)

// 	updater := NewStatusUpdater(storage)
// 	sendOrderToJobs := NewOrderToAccuralSys(order)
// 	OrderJob := make(chan OrderToAccuralSys)
// 	logger.InfoLogger("Activated GetStatusFromAccural")

// 	result := make(chan OrderUpdateFromAccural)
// 	// defer close(result)
// 	for w := 1; w <= 5; w++ {
// 		logger.InfoLogger("Activate workers")
// 		go worker(OrderJob, result)
// 	}

// 	OrderJob <- sendOrderToJobs
// 	defer close(OrderJob)

// 	go func() {
// 		select {
// 		case orderToUpdate := <-result:

// 			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// 			updater.s.UpdateStatus(ctx, orderToUpdate)

// 			cancel()
// 		}
// 	}()

// 	return nil
// }
