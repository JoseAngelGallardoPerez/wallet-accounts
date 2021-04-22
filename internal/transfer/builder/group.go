package builder

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/shopspring/decimal"
)

// Group is a list of actions
type Group []transfer.Action

// Sum summarizes all amounts in the group
func (g Group) Sum() decimal.Decimal {
	var sum decimal.Decimal
	for _, v := range g {
		sum = sum.Add(v.Amount().Mul(decimal.NewFromInt(int64(v.Sign()))))
	}
	return sum
}
