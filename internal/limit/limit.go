package limit

import (
	"github.com/shopspring/decimal"
)

var zero = decimal.NewFromInt(0)

// Limit is used in order to match any value with a certain limit
type Limit interface {
	Available() Value
	WithinLimit(amount CurrencyAmount) error
}

// CurrencyAmount specifies limit amount in specified currency
type CurrencyAmount interface {
	Amount() decimal.Decimal
	CurrencyCode() string
}

// Value defines limit value
type Value interface {
	NoLimit() bool
	CurrencyAmount() CurrencyAmount
}

type amount struct {
	currencyCode string
	amount       decimal.Decimal
}

func Amount(available decimal.Decimal, currencyCode string) CurrencyAmount {
	return &amount{
		currencyCode: currencyCode,
		amount:       available,
	}
}

func (a *amount) Amount() decimal.Decimal {
	return a.amount
}

func (a *amount) CurrencyCode() string {
	return a.currencyCode
}
