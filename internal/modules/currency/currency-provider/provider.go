package currency_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/currency"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/connection"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/service"
)

func Providers() []interface{} {
	return []interface{}{
		connection.NewCurrencyConnection,
		serializer.NewCurrencySerializer,
		service.NewCurrenciesService,
		currency.NewProvider,
		currency.NewRateSource,
	}
}
