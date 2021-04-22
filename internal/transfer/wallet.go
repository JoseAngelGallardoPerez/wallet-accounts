package transfer

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// Debitable represents an instance that cold be debited
type Debitable interface {
	CurrencyAmount
	// Debit debits funds from the debitable
	Debit(amount CurrencyAmount) error
}

// Debitable represents an instance that cold be credited
type Creditable interface {
	// Currency returns currency
	Currency() Currency
	// CreditFromAlias credits funds to the creditable
	Credit(amount CurrencyAmount) error
}

// Wallet represents mix of balance and currency which could be used in order to control operations
// with different currencies
type Wallet struct {
	balance  Balance
	currency Currency
}

// NewWallet is a Wallet constructor
func NewWallet(balance Balance, currency Currency) *Wallet {
	return &Wallet{balance: balance, currency: currency}
}

// String returns wallet string representation e.g. "EUR 123.45"
func (w *Wallet) String() string {
	return fmt.Sprintf(
		"%s %s",
		w.currency.Code(),
		w.balance.Amount().StringFixed(int32(w.currency.fraction)),
	)
}
func (w *Wallet) Currency() Currency {
	return w.currency
}

// GetCurrencyCode returns wallet currency code
func (w *Wallet) CurrencyCode() string {
	return w.currency.Code()
}

// Amount returns wallet balance value
func (w *Wallet) Amount() decimal.Decimal {
	return w.balance.Amount()
}

// CreditFromAlias credits funds to the wallet balance
func (w *Wallet) Credit(amount CurrencyAmount) error {
	amountCur := amount.Currency()
	if w.currency.Code() != amountCur.Code() {
		return errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be credited to the wallet: want %s got %s",
			w.currency.Code(),
			amountCur.Code(),
		)
	}
	if amount.Amount().LessThanOrEqual(decimal.New(0, 1)) {
		return errors.Wrapf(
			ErrInvalidAmount,
			"expected the amount greater than 0, got %s",
			amount.Amount().String(),
		)
	}
	return w.balance.Add(amount.Amount())
}

// Debit debits funds from the wallet balance
func (w *Wallet) Debit(amount CurrencyAmount) error {
	amountCur := amount.Currency()
	if w.currency.Code() != amountCur.Code() {
		return errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be debited from the wallet: want %s got %s",
			w.currency.Code(),
			amountCur.Code(),
		)
	}
	if amount.Amount().LessThanOrEqual(decimal.New(0, 1)) {
		return errors.Wrapf(
			ErrInvalidAmount,
			"expected the amount greater than 0, got %s",
			amount.Amount().String(),
		)
	}
	return w.balance.Sub(amount.Amount())
}

// NoOpWallet is used in order to perform debit or credit operations without changing any balance
type NoOpWallet struct {
	currency Currency
}

func NewNoOpWallet(currency Currency) *NoOpWallet {
	return &NoOpWallet{currency: currency}
}

// Credit simulates credit operation
func (n NoOpWallet) Credit(amount CurrencyAmount) error {
	amountCur := amount.Currency()
	if n.currency.Code() != amountCur.Code() {
		return errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be debited from the wallet: want %s got %s",
			n.currency.Code(),
			amountCur.Code(),
		)
	}
	if amount.Amount().LessThanOrEqual(decimal.New(0, 1)) {
		return errors.Wrapf(
			ErrInvalidAmount,
			"expected the amount greater than 0, got %s",
			amount.Amount().String(),
		)
	}
	return nil
}

// Currency return given currency
func (n NoOpWallet) Currency() Currency {
	return n.currency
}

// Amount always returns 0
func (n NoOpWallet) Amount() decimal.Decimal {
	return decimal.NewFromInt(0)
}

// Debit simulates debit operation
func (n NoOpWallet) Debit(amount CurrencyAmount) error {
	amountCur := amount.Currency()
	if n.currency.Code() != amountCur.Code() {
		return errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be debited from the wallet: want %s got %s",
			n.currency.Code(),
			amountCur.Code(),
		)
	}
	if amount.Amount().LessThanOrEqual(decimal.New(0, 1)) {
		return errors.Wrapf(
			ErrInvalidAmount,
			"expected the amount greater than 0, got %s",
			amount.Amount().String(),
		)
	}
	return nil
}
