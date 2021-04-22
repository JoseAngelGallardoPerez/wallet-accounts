package transfer

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"

	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// ExchangeAction is used in order to exchange one currency to another
type ExchangeAction struct {
	rateSource exchange.RateSource
	amount     CurrencyAmount
	toCurrency Currency
	result     decimal.Decimal
	performed  bool
}

// NewExchangeAction is ExchangeAction constructor
func NewExchangeAction(amount CurrencyAmount, rateSource exchange.RateSource, toCurrency Currency) Action {
	return &ExchangeAction{rateSource: rateSource, amount: amount, toCurrency: toCurrency}
}

// Currency returns resulting currency
func (e *ExchangeAction) Currency() Currency {
	return e.toCurrency
}

// Sign return indicates whether action is credit or debit
func (e *ExchangeAction) Sign() int {
	return 0
}

// Amount returns converted amount. It always returns 0 until action is performed.
func (e *ExchangeAction) Amount() decimal.Decimal {
	return e.result
}

// Perform executes currency conversion
func (e *ExchangeAction) Perform() error {
	if e.performed {
		return errors.Wrapf(
			ErrAlreadyPerformed,
			"exchange action \"%s/%s\" has been already performed",
			e.Purpose(),
			e.Message(),
		)
	}
	e.performed = true
	fromCur, toCur := e.amount.Currency(), e.toCurrency
	rate, err := e.rateSource.FindRate(fromCur.Code(), toCur.Code())
	if err != nil {
		return errors.Wrapf(err, "failed to perform exchange currency action: %s", err.Error())
	}
	e.result = e.amount.Amount().Mul(rate.Rate())
	return nil
}

// IsPerformed indicates whether the action is performed
func (e *ExchangeAction) IsPerformed() bool {
	return e.performed
}

// Purpose indicates action purpose
func (e *ExchangeAction) Purpose() string {
	return "exchange currency"
}

// Message shows additional details of the action
func (e *ExchangeAction) Message() string {
	fromCur, toCur := e.amount.Currency(), e.toCurrency
	return fmt.Sprintf("from %s to %s", fromCur.Code(), toCur.Code())
}
