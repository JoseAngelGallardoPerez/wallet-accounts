package service

import "github.com/shopspring/decimal"

type Rate struct {
	Rate           decimal.Decimal
	ExchangeMargin decimal.Decimal
}
