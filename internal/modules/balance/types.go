package balance

import (
	"github.com/shopspring/decimal"
)

type Definer interface {
	// TypeName returns corresponding balance type name e.g. account, card, revenue_account
	TypeName() string
	// GetId returns unique id for the given balance type if possible
	GetId() *uint64
	// GetUserId returns owner user id if possible, nil must be returned if balance has no owner or owned by the system
	GetUserId() *string
	// GetCurrencyCode returns balance related currency
	GetCurrencyCode() (string, error)
}

type Balance interface {
	Definer
	// CurrentBalance retrieves actual balance information
	CurrentBalance() (decimal.Decimal, error)
	// AvailableBalance calculates available balance value
	AvailableBalance() (decimal.Decimal, error)
}

type Difference struct {
	// BalanceType such as account, card, revenue_account
	BalanceType string `json:"balanceType"`
	// BalanceId represents balance id if exists
	BalanceId *uint64 `json:"balanceId"`
	//GetCurrencyCode such as EUR
	CurrencyCode string `json:"currencyCode"`
	// Difference signed number which represents debit or credit operation on the balance
	Difference decimal.Decimal `json:"difference"`
}
