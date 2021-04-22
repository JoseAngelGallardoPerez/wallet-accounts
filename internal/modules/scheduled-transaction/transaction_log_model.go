package scheduled_transaction

import (
	"time"

	"github.com/shopspring/decimal"
)

type ScheduledTransactionLog struct {
	Id                     *uint64               `json:"id"`
	ScheduledTransactionId *uint64               `json:"scheduledTransactionId"`
	ScheduledTransaction   *ScheduledTransaction `json:"scheduledTransaction"`
	Amount                 decimal.Decimal       `json:"amount"`
	CreatedAt              *time.Time            `json:"createdAt"`
}
