package getbonuses

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/logger"
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

func (storage *PsqURLlStorage) UpdateStatus(ctx context.Context, order OrderUpdateFromAccural, login string) error {

	_, err := storage.db.ExecContext(ctx, `UPDATE orders
		SET status = $1, accrual = $2
		WHERE "order" = $3`, order.Status, order.Accrual, order.Order)
	if err != nil {
		logger.ErrorLogger("Error making an update request", err)
	}

	_, err = storage.db.ExecContext(ctx, `UPDATE users
		SET balance = balance + $1
		WHERE login = $2`, order.Accrual, login)
	if err != nil {
		logger.ErrorLogger("Error making an update request", err)
	}
	return nil
}

func GetStatusFromAccural(order string, login string) <-chan OrderUpdateFromAccural {

	var wg sync.WaitGroup

	sendOrderToJobs := NewOrderToAccuralSys(order)
	OrderJob := make(chan OrderToAccuralSys)
	result := make(chan OrderUpdateFromAccural)

	defer close(result)

	wg.Add(1)
	go func(jobs <-chan OrderToAccuralSys, result chan<- OrderUpdateFromAccural) {
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

	}(OrderJob, result)

	OrderJob <- sendOrderToJobs
	defer close(OrderJob)

	wg.Wait()

	return result
}
