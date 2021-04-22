package transfer

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// CreditAction is used in order to perform debit action on some debitable instance
type DebitAction struct {
	debitable Debitable
	amount    CurrencyAmount
	performed bool
	purpose   string
	message   string
}

// NewDebitAction is CreditAction constructor
func NewDebitAction(debitable Debitable, amount CurrencyAmount) (*DebitAction, error) {
	amountCur := amount.Currency()
	debitCur := debitable.Currency()
	if amountCur.Code() != debitCur.Code() {
		return nil, errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be debited: want %s got %s",
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
	return &DebitAction{
		debitable: debitable,
		amount:    amount,
		performed: false,
		purpose:   "debit",
		message:   "",
	}, nil
}

// Sign indicates whether action is credit or debit
func (d *DebitAction) Sign() int {
	return -1
}

// Currency returns debit currency
func (d *DebitAction) Currency() Currency {
	return d.debitable.Currency()
}

// Amount returns debited amount.
// Note that if action is not performed then Amount returns zero
func (d *DebitAction) Amount() decimal.Decimal {
	if d.performed {
		return d.amount.Amount()
	}
	return decimal.NewFromInt(0)
}

// Perform performs debit action
func (d *DebitAction) Perform() error {
	if d.performed {
		return errors.Wrapf(
			ErrAlreadyPerformed,
			"debit action \"%s/%s\" has already been performed",
			d.purpose,
			d.message,
		)
	}
	d.performed = true
	// ignore debit if value is zero
	if d.amount.Amount().IsZero() {
		return nil
	}
	return d.debitable.Debit(d.amount)
}

// IsPerformed determines whether debit action was performed
func (d *DebitAction) IsPerformed() bool {
	return d.performed
}

// Purpose describes action purpose
func (d *DebitAction) Purpose() string {
	return d.purpose
}

// Message is optional piece of information that clarifies the action
func (d *DebitAction) Message() string {
	return d.message
}

// SetMessage sets custom message
func (d *DebitAction) SetMessage(message string) {
	d.message = message
}

// SetPurpose sets custom purpose
func (d *DebitAction) SetPurpose(purpose string) {
	d.purpose = purpose
}
