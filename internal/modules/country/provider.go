package country

import (
	"github.com/Confialink/wallet-accounts/internal/modules/country/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/country/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/country/service"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewCountryRepository,
		service.NewCountryService,

		handler.NewCountryHandler,
	}
}
