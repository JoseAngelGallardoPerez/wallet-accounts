package limit

import (
	"github.com/shopspring/decimal"
)

type val struct {
	amount CurrencyAmount
}

func Val(amount decimal.Decimal, currencyCode string) Value {
	return &val{
		amount: Amount(amount, currencyCode),
	}
}

func (v *val) Amount() decimal.Decimal {
	return v.amount.Amount()
}

func (v *val) CurrencyCode() string {
	return v.amount.CurrencyCode()
}

func (v *val) NoLimit() bool {
	return false
}

func (v *val) CurrencyAmount() CurrencyAmount {
	return v
}
