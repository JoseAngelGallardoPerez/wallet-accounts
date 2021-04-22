package account_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	"github.com/Confialink/wallet-accounts/internal/modules/account/wrapper"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewAccountRepository,
		repository.NewRevenueAccountRepository,
		service.NewCsv,
		service.NewAccountService,
		service.NewRevenueAccountService,
		serializer.NewAccountSerializer,
		wrapper.NewAccountCreator,

		handler.NewAccountHandler,
		handler.NewCsvHandler,
		handler.NewRevenueAccountHandler,
		handler.NewHandlerParams,
	}
}
