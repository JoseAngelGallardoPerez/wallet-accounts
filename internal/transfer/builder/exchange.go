package builder

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"fmt"
)

// ExchangeUsing declares exchange dialog options
type ExchangeUsing interface {
	// Using sets exchange rates source
	Using(source exchange.RateSource) ExchangeTo
}

// ExchangeTo declares exchange dialog options
type ExchangeTo interface {
	// ToCurrency specifies destination currency
	ToCurrency(interface{}) *Transfer
}

type exchangeTo struct {
	using *exchangeUsing
}

// ToCurrency specifies destination currency
// in case if currency is string it treated as alias name
// it could also be transfer.CurrencyAmount, transfer.Creditable or transfer.Currency
func (e *exchangeTo) ToCurrency(currency interface{}) *Transfer {
	var destinationCurrency transfer.Currency

	switch currency := currency.(type) {
	case string:
		destinationCurrency = e.findCurrencyAlias(currency)
		if destinationCurrency.Code() == "" {
			panic(fmt.Sprintf("unable to find alias by name %s", currency))
		}
	case transfer.Creditable:
		destinationCurrency = currency.Currency()
	case transfer.CurrencyAmount:
		destinationCurrency = currency.Currency()
	case transfer.Currency:
		destinationCurrency = currency
	case *transfer.Currency:
		destinationCurrency = *currency
	default:
		panic(fmt.Sprintf("failed to create exchange action: unexpected amount type is given %T", currency))
	}

	t := e.using.transfer
	exchangeAction := transfer.NewExchangeAction(e.using.fromCurrAmount, e.using.rateSource, destinationCurrency)
	t.actions = append(t.actions, exchangeAction)
	t.contextAction = exchangeAction
	return t
}

func (e *exchangeTo) findCurrencyAlias(alias string) transfer.Currency {
	t := e.using.transfer
	//if cred, ok := t.credAliases[alias]; ok {
	//	return cred.Currency()
	//}
	if amount, ok := t.amountAliases[alias]; ok {
		return amount.Currency()
	}
	return transfer.Currency{}
}

type exchangeUsing struct {
	rateSource     exchange.RateSource
	transfer       *Transfer
	fromCurrAmount transfer.CurrencyAmount
}

// Using sets exchange rates source
func (e *exchangeUsing) Using(source exchange.RateSource) ExchangeTo {
	e.rateSource = source
	return &exchangeTo{
		using: e,
	}
}

// Exchange creates new transfer by starting exchange dialog
func Exchange(amount transfer.CurrencyAmount) ExchangeUsing {
	return exchangeU(amount, New())
}

func exchangeU(amount transfer.CurrencyAmount, t *Transfer) ExchangeUsing {
	return &exchangeUsing{
		transfer:       t,
		fromCurrAmount: amount,
	}
}
