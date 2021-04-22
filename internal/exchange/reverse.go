package exchange

import (
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// reverseRateSource calculates source using reverse rate
// example: a/b = x -> b/a = 1/x
type reverseRateSource struct {
	rateSource RateSource
}

// NewReverseRateSource creates reverse rate source
func NewReverseRateSource(source RateSource) RateSource {
	return &reverseRateSource{
		rateSource: source,
	}
}

// FindRate tries to find rate using the given source.
// In case if no rate found it tries to find reverse rate and calculate required rate with it.
// a/b = x -> b/a = 1/x
func (r *reverseRateSource) FindRate(base, reference string) (Rate, error) {
	originalRate, err := r.rateSource.FindRate(base, reference)
	if err == nil {
		return originalRate, nil
	}
	if errors.Cause(err) != ErrRateNotFound {
		return Rate{}, errors.Wrapf(err, "failed to find reverse rate %s -> %s", base, reference)
	}

	reverseRate, err := r.rateSource.FindRate(reference, base)
	if err != nil {
		return Rate{}, errors.Wrapf(err, "failed to find reverse rate %s -> %s", base, reference)
	}
	return Rate{
		base:      base,
		reference: reference,
		rate:      decimal.NewFromInt(1).Div(reverseRate.Rate()),
	}, nil
}
