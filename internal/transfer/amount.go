package transfer

import (
	"github.com/Confialink/wallet-accounts/pkg/decround"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// CurrencyValue represents types that contain both currency and amount
type CurrencyAmount interface {
	// Currency returns currency
	Currency() Currency
	// Amount returns value representation
	Amount() decimal.Decimal
}

// Amountable represents types that could provide amounts
type Amountable interface {
	// Amount returns value representation
	Amount() decimal.Decimal
}

// Amount is a basic container for a currency and amount
type Amount struct {
	currency Currency
	amount   decimal.Decimal
}

// NewAmount is Amount constructor
func NewAmount(currency Currency, amount decimal.Decimal) CurrencyAmount {
	return &Amount{currency: currency, amount: amount}
}

// Currency returns currency
func (a *Amount) Currency() Currency {
	return a.currency
}

// Amount returns amount
func (a *Amount) Amount() decimal.Decimal {
	return a.amount
}

// AmountMultiplier changes original amount by multiplying it by the multiplier
type AmountMultiplier struct {
	topAmount  CurrencyAmount
	multiplier decimal.Decimal
}

// NewAmountMultiplier is AmountMultiplier constructor
func NewAmountMultiplier(topAmount CurrencyAmount, multiplier decimal.Decimal) CurrencyAmount {
	return &AmountMultiplier{topAmount: topAmount, multiplier: multiplier}
}

// Currency returns currency
func (a *AmountMultiplier) Currency() Currency {
	return a.topAmount.Currency()
}

// Amount returns original amount multiplied by the multiplier
func (a *AmountMultiplier) Amount() decimal.Decimal {
	return a.topAmount.Amount().Mul(a.multiplier)
}

// AmountConsumable is amount wrapper that can be debited
type AmountConsumable struct {
	topAmount CurrencyAmount
	spent     decimal.Decimal
}

// NewAmountConsumable is AmountConsumable constructor
func NewAmountConsumable(topAmount CurrencyAmount) *AmountConsumable {
	return &AmountConsumable{
		topAmount: topAmount,
		spent:     decimal.New(0, 1),
	}
}

// Debit debits funds from the amount
func (a *AmountConsumable) Debit(amount CurrencyAmount) error {
	amountCur := amount.Currency()
	cur := a.topAmount.Currency()
	if cur.Code() != amountCur.Code() {
		return errors.Wrapf(
			ErrCurrenciesMismatch,
			"only the amount in the same currency can be spent: want %s got %s",
			cur.Code(),
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
	remains := a.Amount()
	if remains.LessThan(amount.Amount()) {
		return errors.Wrapf(
			ErrNotEnoughFunds,
			"not enough funds to spend requested %s got %s",
			amount.Amount().String(),
			remains.String(),
		)
	}
	a.spent = a.spent.Add(amount.Amount())
	return nil
}

// Currency returns currency
func (a *AmountConsumable) Currency() Currency {
	return a.topAmount.Currency()
}

// Amount returns remaining amount
func (a *AmountConsumable) Amount() decimal.Decimal {
	return a.topAmount.Amount().Sub(a.spent)
}

// roundAmountGeneral is general part of rounded amount
type roundAmountGeneral struct {
	topAmount CurrencyAmount
	roundFunc decround.Func
}

// Currency returns top amount currency
func (r *roundAmountGeneral) Currency() Currency {
	return r.topAmount.Currency()
}

// Amount returns rounded amount
func (r *roundAmountGeneral) Amount() decimal.Decimal {
	currency := r.topAmount.Currency()
	return r.roundFunc(r.topAmount.Amount(), int32(currency.Fraction()))
}

// roundAmountRemainder is a remainder part of rounded amount
type roundAmountRemainder struct {
	topAmount CurrencyAmount
	roundFunc decround.Func
}

// Currency returns top amount currency
func (r *roundAmountRemainder) Currency() Currency {
	return r.topAmount.Currency()
}

// Amount returns the remainder that is difference between the given amount and the rounded
func (r *roundAmountRemainder) Amount() decimal.Decimal {
	currency := r.topAmount.Currency()
	trunc := r.roundFunc(r.topAmount.Amount(), int32(currency.Fraction()))
	return r.topAmount.Amount().Sub(trunc)
}

// NewRoundAmount creates rounded amount from the given amount to the given currency fraction precision
// it returns the rounded amount and remainder that is difference between the given amount and the rounded
func NewRoundAmount(amount CurrencyAmount, roundFunc decround.Func) (CurrencyAmount, CurrencyAmount) {
	general := &roundAmountGeneral{
		topAmount: amount,
		roundFunc: roundFunc,
	}
	remainder := &roundAmountRemainder{
		topAmount: amount,
		roundFunc: roundFunc,
	}
	return general, remainder
}

// AmountAbs is used in order to decorate top amount as absolute
type AmountAbs struct {
	topAmount CurrencyAmount
}

// NewAmountAbs is AmountAbs constructor
func NewAmountAbs(topAmount CurrencyAmount) *AmountAbs {
	return &AmountAbs{topAmount: topAmount}
}

// Currency returns the given currency
func (a *AmountAbs) Currency() Currency {
	return a.topAmount.Currency()
}

// Amount returns absolute amount
func (a *AmountAbs) Amount() decimal.Decimal {
	return a.topAmount.Amount().Abs()
}

// AmountNeg is used in order to negotiate given amount
type AmountNeg struct {
	topAmount CurrencyAmount
}

// NewAmountNeg is AmountNeg constructor
func NewAmountNeg(topAmount CurrencyAmount) *AmountNeg {
	return &AmountNeg{topAmount: topAmount}
}

// Currency returns top amount currency
func (a *AmountNeg) Currency() Currency {
	return a.topAmount.Currency()
}

// Amount returns negotiated amount
func (a *AmountNeg) Amount() decimal.Decimal {
	return a.topAmount.Amount().Neg()
}
