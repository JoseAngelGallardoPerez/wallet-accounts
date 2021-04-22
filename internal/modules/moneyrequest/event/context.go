package event

import "github.com/shopspring/decimal"

type Context struct {
	MoneyRequestId  uint64
	RecipientUID    string
	Amount          decimal.Decimal
	Currency        string
	SenderFirstName string
	SenderLastName  string
}
