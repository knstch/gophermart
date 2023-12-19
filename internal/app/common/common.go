package common

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/knstch/gophermart/cmd/config"
	"github.com/knstch/gophermart/internal/app/logger"
)

// A function that sends orders to accrual system
// and receives status updates with accrued bonuses.
func GetStatusFromAccrual(order Order) OrderUpdateFromAccural {
	client := resty.New().SetBaseURL(config.ReadyConfig.Accural)
	var orderUpdate OrderUpdateFromAccural
	for {
		resp, err := client.R().
			SetResult(&orderUpdate).
			Get("/api/orders/" + order.Order)
		if err != nil {
			logger.ErrorLogger("Got error trying to send a get request to accrual: ", err)
			break
		}
		switch resp.StatusCode() {
		case 429:
			time.Sleep(3 * time.Second)
		case 204:
			time.Sleep(1 * time.Second)
		}

		if resp.StatusCode() == 500 {
			logger.ErrorLogger("Internal server error in accrual system: ", err)
			break
		}

		if orderUpdate.Status == "INVALID" || orderUpdate.Status == "PROCESSED" {
			break
		}
	}
	return orderUpdate
}
