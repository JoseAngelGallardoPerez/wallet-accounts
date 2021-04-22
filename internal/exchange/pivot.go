package exchange

import "github.com/pkg/errors"

// pivotRateSource is used in order to calculate rate with "pivot" currency
// for example("a" is pivot): a/b = x; a/c = y; b/c = x/y
type pivotRateSource struct {
	pivotCurrencyCode string
	ratesSource       RateSource
}

// NewPivotRateSource is pivot rate source constructor
func NewPivotRateSource(pivotCurrencyCode string, source RateSource) RateSource {
	return &pivotRateSource{
		pivotCurrencyCode: pivotCurrencyCode,
		ratesSource:       source,
	}
}

// FindRate fetches rate from the given source. If not found it tries to calculate the rate.
func (p *pivotRateSource) FindRate(base, reference string) (Rate, error) {
	if base == p.pivotCurrencyCode {
		return p.ratesSource.FindRate(base, reference)
	}
	r, err := p.indirectRate(base, reference)
	if err != nil {
		return r, errors.Wrapf(err, "failed to calculate pivot rate for %s -> %s", base, reference)
	}
	return r, nil
}

func (p *pivotRateSource) indirectRate(base, reference string) (Rate, error) {
	nilRate := Rate{}
	baseRate, err := p.ratesSource.FindRate(p.pivotCurrencyCode, base)
	if err != nil {
		return nilRate, err
	}
	referenceRate, err := p.ratesSource.FindRate(p.pivotCurrencyCode, reference)
	if err != nil {
		return nilRate, err
	}
	ratio := referenceRate.Rate().Div(baseRate.Rate())
	return Rate{
		base:      base,
		reference: reference,
		rate:      ratio,
	}, nil
}
