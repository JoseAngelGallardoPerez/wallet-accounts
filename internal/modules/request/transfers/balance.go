package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// SufficientBalancePermission is simple rule that checks balance
type SufficientBalancePermission struct {
	requested transfer.Amountable
	available transfer.Amountable
}

// NewSufficientBalancePermission is SufficientBalancePermission constructor
func NewSufficientBalancePermission(requested transfer.Amountable, available transfer.Amountable) *SufficientBalancePermission {
	return &SufficientBalancePermission{requested: requested, available: available}
}

// Check checks whether available amount can cover requested amount
func (s *SufficientBalancePermission) Check() error {
	requestedAmount := s.requested.Amount()
	availableAmount := s.available.Amount()
	if requestedAmount.GreaterThan(availableAmount) {
		return errors.Wrapf(
			ErrInsufficientBalance,
			`requested amount "%s" is greater than available amount "%s"`,
			requestedAmount.String(),
			availableAmount.String(),
		)
	}
	return nil
}

func (s *SufficientBalancePermission) Name() string {
	return "sufficient_balance"
}

// SimpleAmountable is a wrapper around decimal
type SimpleAmountable decimal.Decimal

// Amount returns itself
func (s SimpleAmountable) Amount() decimal.Decimal {
	return decimal.Decimal(s)
}
