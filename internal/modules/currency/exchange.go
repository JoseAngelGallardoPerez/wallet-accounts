package currency

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	currenciesService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/pkg/errors"
	"strings"
)

type rateSource struct {
	currenciesService currenciesService.CurrenciesServiceInterface
}

func NewRateSource(currenciesService currenciesService.CurrenciesServiceInterface) exchange.RateSource {
	return &rateSource{currenciesService: currenciesService}
}

func (r *rateSource) FindRate(base, reference string) (exchange.Rate, error) {
	rate, err := r.currenciesService.GetCurrenciesRateByCodes(base, reference)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			err = ErrExchangeRateNotFound
		}
		return exchange.Rate{}, errors.Wrapf(
			err,
			"failed to get rate %s/%s",
			base,
			reference,
		)
	}
	return exchange.NewRate(base, reference, rate.Rate), nil
}
