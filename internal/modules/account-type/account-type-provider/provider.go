package account_type_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/service"
)

func Providers() []interface{} {
	return []interface{}{
		service.NewAccountTypeService,
		repository.NewAccountTypeRepository,

		handler.NewAccountTypeHandler,
	}
}
