package exchange

import "github.com/shopspring/decimal"

// Rate defines exchange rate between base and reference codes
type Rate struct {
	base      string
	reference string
	rate      decimal.Decimal
}

// NewRate is Rate constructor
func NewRate(base string, reference string, rate decimal.Decimal) Rate {
	return Rate{
		base:      base,
		reference: reference,
		rate:      rate,
	}
}

// BaseCurrencyCode returns base currency code
func (r *Rate) BaseCurrencyCode() string {
	return r.base
}

// ReferenceCurrencyCode returns reference currency code
func (r *Rate) ReferenceCurrencyCode() string {
	return r.reference
}

// Rate returns current rate value
func (r *Rate) Rate() decimal.Decimal {
	return r.rate
}
