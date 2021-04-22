package transfer

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// CreditAction is used in order to perform debit action on some creditable instance
type CreditAction struct {
	creditable Creditable
	amount     CurrencyAmount
	performed  bool
	purpose    string
	message    string
}

// NewDebitAction is CreditAction constructor
func NewCreditAction(creditable Creditable, amount CurrencyAmount) (*CreditAction, error) {
	amountCur := amount.Currency()
	debitCur := creditable.Currency()
	if amountCur.Code() != debitCur.Code() {
		return nil, errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be credited: want %s got %s",
			debitCur.Code(),
			amountCur.Code(),
		)
	}
	if amount.Amount().LessThan(decimal.New(0, 1)) {
		return nil, errors.Wrapf(
			ErrInvalidAmount,
			"expected the amount greater than or equal to 0, got %s",
			amount.Amount().String(),
		)
	}
	return &CreditAction{
		creditable: creditable,
		amount:     amount,
		performed:  false,
		purpose:    "credit",
		message:    "",
	}, nil
}

// Currency returns debit currency
func (d *CreditAction) Currency() Currency {
	return d.creditable.Currency()
}

// Amount returns debited amount.
// Note that if action is not performed then Amount returns zero
func (d *CreditAction) Amount() decimal.Decimal {
	if d.performed {
		return d.amount.Amount()
	}
	return decimal.NewFromInt(0)
}

// Sign indicates whether action is credit or debit
func (d *CreditAction) Sign() int {
	return 1
}

// Perform performs credit action
func (d *CreditAction) Perform() error {
	if d.performed {
		return errors.Wrapf(
			ErrAlreadyPerformed,
			"credit action \"%s/%s\" has already been performed",
			d.purpose,
			d.message,
		)
	}
	d.performed = true
	// ignore debit if value is zero
	if d.amount.Amount().IsZero() {
		return nil
	}
	return d.creditable.Credit(d.amount)
}

// IsPerformed determines whether credit action was performed
func (d *CreditAction) IsPerformed() bool {
	return d.performed
}

// Purpose describes action purpose
func (d *CreditAction) Purpose() string {
	return d.purpose
}

// Message is optional piece of information that clarifies the action
func (d *CreditAction) Message() string {
	return d.message
}

// SetMessage sets custom message
func (d *CreditAction) SetMessage(message string) {
	d.message = message
}

// SetPurpose sets custom purpose
func (d *CreditAction) SetPurpose(purpose string) {
	d.purpose = purpose
}
