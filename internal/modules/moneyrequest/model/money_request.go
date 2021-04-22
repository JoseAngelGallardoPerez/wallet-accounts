package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
)

type MoneyRequest struct {
	ID                 uint64          `json:"id"`
	TargetUserID       string          `json:"targetUID"` // ID of user who will receive a request to make TBU transfer
	Sender             *User           `json:"sender,omitempty" gorm:"-"`
	InitiatorUserID    string          `json:"initiatorUID"` // Initiator of MoneyRequest. This user will receive funds from another user
	Recipient          *User           `json:"recipient,omitempty" gorm:"-"`
	Status             string          `json:"status"`
	RecipientAccountID uint64          `json:"recipientAccountId"`
	RequestID          *uint64         `json:"requestId"`
	Amount             decimal.Decimal `json:"amount"`
	CurrencyCode       string          `json:"currencyCode"`
	Description        string          `json:"description"`
	IsNew              bool            `json:"isNew"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
}

type User struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
}

func (*MoneyRequest) TableName() string {
	return "money_requests"
}
