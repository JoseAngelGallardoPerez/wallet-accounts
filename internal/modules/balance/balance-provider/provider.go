package balance_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-accounts/internal/modules/balance/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/balance/service"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewType,
		repository.NewSnapshot,
		service.NewSnapshot,
		balance.NewDefaultResolver,
		balance.NewDefaultReducer,
		balance.NewDBAggregationFactory,
		balance.NewAggregationService,
	}
}
