package model

import (
	"time"

	"github.com/Confialink/wallet-accounts/internal/modules/currency/model"
	"github.com/shopspring/decimal"
)

// TableName sets RevenueAccount's table name to be `revenue_accounts`
func (RevenueAccountModel) TableName() string {
	return "revenue_accounts"
}

// RevenueAccountModel revenue account model
type RevenueAccountModel struct {
	RevenueAccountPublic
	RevenueAccountPrivate
}

// RevenueAccountPublic is used to create a new account
type RevenueAccountPublic struct {
	Balance      decimal.Decimal `json:"balance"`
	CurrencyCode string          `json:"currencyCode"`
	Currency     *model.Currency `gorm:"-" json:"currency"`
	IsDefault    bool            `json:"isDefault"`
}

// RevenueAccountPrivate contains fields assigned automatically
type RevenueAccountPrivate struct {
	ID              uint64          `gorm:"primary_key" json:"id"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	AvailableAmount decimal.Decimal `json:"availableAmount"`
}

func (a *RevenueAccountModel) BeforeCreate() {
	a.AvailableAmount = a.Balance
}

func (a *RevenueAccountModel) CurrentBalance() (decimal.Decimal, error) {
	return a.Balance, nil
}

func (a *RevenueAccountModel) AvailableBalance() (decimal.Decimal, error) {
	return a.AvailableAmount, nil
}

func (a *RevenueAccountModel) GetCurrencyCode() (string, error) {
	return a.CurrencyCode, nil
}

func (a *RevenueAccountModel) TypeName() string {
	return "revenue_account"
}

func (a *RevenueAccountModel) GetId() *uint64 {
	return &a.ID
}

func (a *RevenueAccountModel) GetUserId() *string {
	return nil
}
