package limit

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// Max creates "max" limit
func Max(amount decimal.Decimal, currencyCode string) Limit {
	return &max{
		amount:       amount,
		currencyCode: currencyCode,
	}
}

// max represents limited amount
type max struct {
	amount       decimal.Decimal
	currencyCode string
}

// NoLimit always returns false because amount is limited
func (m *max) NoLimit() bool {
	return false
}

// CurrencyAmount returns max itself
func (m *max) CurrencyAmount() CurrencyAmount {
	return m
}

// max retrieves available amount
func (m *max) Amount() decimal.Decimal {
	return m.amount
}

// CurrencyCode returns currency code
func (m *max) CurrencyCode() string {
	return m.currencyCode
}

func (m *max) Available() Value {
	return m
}

func (m *max) WithinLimit(amount CurrencyAmount) error {
	currencyCode := m.currencyCode
	available := m.amount

	if amount.Amount().LessThanOrEqual(zero) {
		return errors.Wrapf(
			ErrInvalidAmount,
			"expected value to be greater than 0, got %s",
			amount.Amount().String(),
		)
	}
	if currencyCode != amount.CurrencyCode() {
		return errors.Wrapf(
			ErrCurrenciesMismatch,
			"requested amount currency %s does not match limit currency %s",
			amount.CurrencyCode(),
			currencyCode,
		)
	}
	if available.LessThan(amount.Amount()) {
		return errors.Wrapf(
			ErrLimitExceeded,
			"the requested value %s exceeds the available limit %s",
			amount.Amount().String(),
			available.String(),
		)
	}
	return nil
}
