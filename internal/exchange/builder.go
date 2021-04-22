package exchange

import "github.com/shopspring/decimal"

// RateBuilder helps to create rate
type RateBuilder string

// One accept base currency code and initiate building process
func One(base string) RateBuilder {
	return RateBuilder(base)
}

// Is accept rate value and reference currency.
func (b RateBuilder) Is(rate decimal.Decimal, reference string) Rate {
	return Rate{
		base:      string(b),
		reference: reference,
		rate:      rate,
	}
}
