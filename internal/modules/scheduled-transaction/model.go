package scheduled_transaction

import (
	"time"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/shopspring/decimal"
)

type ScheduledTransaction struct {
	Id            *uint64         `json:"id"`
	Reason        Reason          `json:"reason"`
	AccountId     *uint64         `json:"accountId"`
	Account       *model.Account  `gorm:"foreignkey:AccountId;association_foreignkey:ID" json:"account"`
	Amount        decimal.Decimal `json:"amount"`
	Status        Status          `json:"status"`
	RequestId     *uint64         `json:"requestId"`
	ScheduledDate *time.Time      `json:"scheduledDate"`
	CreatedAt     *time.Time      `json:"createdAt"`
	UpdatedAt     *time.Time      `json:"updatedAt"`
}
