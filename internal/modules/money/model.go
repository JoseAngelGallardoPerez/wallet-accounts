package money

import "github.com/shopspring/decimal"

type Amount struct {
	Value        decimal.Decimal
	CurrencyCode string
}
