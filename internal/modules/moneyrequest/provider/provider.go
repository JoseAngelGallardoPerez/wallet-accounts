package provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/service"
)

// Providers return providers for invoice module
func Providers() []interface{} {
	return []interface{}{
		repository.NewMoneyRequest,
		service.NewMoneyRequest,
		handler.NewMoneyRequest,
		handler.NewParams,
	}
}
