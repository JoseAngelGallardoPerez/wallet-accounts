package form

import (
	"time"

	"github.com/shopspring/decimal"
)

type Account struct {
	Number            string          `json:"number" binding:"required,max=255,alphanum,accountNumberUnique"`
	TypeId            uint64          `json:"typeId" binding:"required"`
	UserId            string          `json:"userId" binding:"required,existUserId,userIsActive"`
	Description       *string         `json:"description"`
	IsActive          *bool           `json:"isActive"`
	InitialBalance    decimal.Decimal `json:"initialBalance"`
	AllowWithdrawals  *bool           `json:"allowWithdrawals"`
	AllowDeposits     *bool           `json:"allowDeposits"`
	MaturityDate      *time.Time      `json:"maturityDate"`
	PayoutDay         *uint64         `json:"payoutDay" binding:"omitempty,gte=1,lte=31"`
	InterestAccountId *uint64         `json:"interestAccountId" binding:"omitempty"`
}
