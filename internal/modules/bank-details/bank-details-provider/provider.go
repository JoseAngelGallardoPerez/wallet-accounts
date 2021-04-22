package bank_details_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/service"
)

func Providers() []interface{} {
	return []interface{}{
		repository.NewAccountRepository,
		service.NewIwtBankAccountService,
		service.NewPdf,

		handler.NewIwtBankAccountHandler,
	}
}
