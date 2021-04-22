package balance

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type Reducer interface {
	Reduce(in AggregationResult, toCurrencyCode string) (out AggregationItem, err error)
}

type defaultReducer struct {
	rateSource exchange.RateSource
}

func NewDefaultReducer(rateSource exchange.RateSource) Reducer {
	return &defaultReducer{rateSource: rateSource}
}

func (d *defaultReducer) Reduce(in AggregationResult, toCurrencyCode string) (out AggregationItem, err error) {
	source := exchange.NewCacheSource(d.rateSource)
	total := decimal.NewFromInt(0)
	for _, r := range in {
		if r.ItemCurrencyCode == toCurrencyCode {
			total = total.Add(r.ItemAmount)
			continue
		}
		rate, err := source.FindRate(r.ItemCurrencyCode, toCurrencyCode)
		if err != nil {
			return AggregationItem{}, errors.Wrapf(
				err,
				"failed to reduce aggregation result because rate source returned error (%s/%s)",
				r.ItemCurrencyCode,
				toCurrencyCode,
			)
		}
		total = total.Add(r.ItemAmount.Mul(rate.Rate()))
	}
	return AggregationItem{
		ItemAmount:       total,
		ItemCurrencyCode: toCurrencyCode,
	}, nil
}
