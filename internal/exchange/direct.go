package exchange

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// directRateSource is simple rate storage
type directRateSource struct {
	rates map[string]map[string]Rate
}

// NewDirectRateSource is constructor for direct rate source
func NewDirectRateSource() RateSourceAndReceiver {
	return &directRateSource{
		rates: make(map[string]map[string]Rate),
	}
}

// Set sets rate
func (d *directRateSource) Set(rate Rate) error {
	if d.rates[rate.BaseCurrencyCode()] == nil {
		d.rates[rate.BaseCurrencyCode()] = make(map[string]Rate)
	}
	d.rates[rate.BaseCurrencyCode()][rate.ReferenceCurrencyCode()] = rate
	return nil
}

// FindRate fetches rate
func (d *directRateSource) FindRate(base, reference string) (Rate, error) {
	if base == reference {
		return Rate{
			base:      base,
			reference: reference,
			rate:      decimal.NewFromInt(1),
		}, nil
	}
	err := errors.Wrapf(ErrRateNotFound, "direct rate not found %s -> %s", base, reference)
	rateMap, ok := d.rates[base]
	if !ok {
		return Rate{}, err
	}
	r, ok := rateMap[reference]
	if !ok {
		return Rate{}, err
	}
	return r, nil
}
