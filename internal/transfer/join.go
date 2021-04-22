package transfer

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// JoinCreditable take 2 creditables and returns 1 joined
func JoinCreditable(c1 Creditable, c2 Creditable) (Creditable, error) {
	cur1, cur2 := c1.Currency(), c2.Currency()
	if cur1.Code() != cur2.Code() {
		return nil, errors.Wrapf(
			ErrCurrenciesMismatch,
			"creditables in different currencies cannot be joined got %s and %s",
			cur1.Code(),
			cur2.Code(),
		)
	}
	return &joinedCreditable{
		c1: c1,
		c2: c2,
	}, nil
}

type joinedCreditable struct {
	c1 Creditable
	c2 Creditable
}

// Currency returns creditable currency
func (j *joinedCreditable) Currency() Currency {
	return j.c1.Currency()
}

// Credit credits both creditables with amount
func (j *joinedCreditable) Credit(amount CurrencyAmount) error {
	err := j.c1.Credit(amount)
	if err != nil {
		return err
	}
	return j.c2.Credit(amount)
}

// JoinDebitable take 2 debitables and returns 1 joined
func JoinDebitable(d1 Debitable, d2 Debitable) (Debitable, error) {
	cur1, cur2 := d1.Currency(), d2.Currency()
	if cur1.Code() != cur2.Code() {
		return nil, errors.Wrapf(
			ErrCurrenciesMismatch,
			"debitables in different currencies cannot be joined got %s and %s",
			cur1.Code(),
			cur2.Code(),
		)
	}
	return &joinedDebitable{
		d1: d1,
		d2: d2,
	}, nil
}

type joinedDebitable struct {
	d1 Debitable
	d2 Debitable
}

// Amount returns smallest amount
func (j *joinedDebitable) Amount() decimal.Decimal {
	a1, a2 := j.d1.Amount(), j.d2.Amount()
	if a1.LessThan(a2) {
		return a1
	}
	return a2
}

// Currency returns debitable currency
func (j *joinedDebitable) Currency() Currency {
	return j.d1.Currency()
}

// Debit debits both debitables with given amount
func (j *joinedDebitable) Debit(amount CurrencyAmount) error {
	err := j.d1.Debit(amount)
	if err != nil {
		return err
	}
	return j.d2.Debit(amount)
}
