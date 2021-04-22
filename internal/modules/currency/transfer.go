package currency

import (
	currenciesService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/transfer"
)

// Provider retrieves currency details with the given service and caches results
type Provider struct {
	cache             map[string]transfer.Currency
	currenciesService currenciesService.CurrenciesServiceInterface
}

// NewProvider is transfer currency provider
func NewProvider(currenciesService currenciesService.CurrenciesServiceInterface) transfer.CurrencyProvider {
	return &Provider{
		currenciesService: currenciesService,
		cache:             map[string]transfer.Currency{},
	}
}

// Get provides currency by code
func (p *Provider) Get(code string) (transfer.Currency, error) {
	if currency, ok := p.cache[code]; ok {
		return currency, nil
	}

	currencyModel, err := p.currenciesService.GetByCode(code)
	if err != nil {
		return transfer.Currency{}, err
	}
	currency := transfer.NewCurrency(code, uint(currencyModel.DecimalPlaces))
	p.cache[code] = *currency
	return *currency, nil
}
