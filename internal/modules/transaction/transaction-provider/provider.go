package transaction_provider

import (
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/http/handler"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/service"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/service/csv"
	transaction_view "github.com/Confialink/wallet-accounts/internal/modules/transaction/transaction-view"
)

func Providers() []interface{} {
	return []interface{}{
		//transaction.TxSerializer
		transaction_view.ProvideDefaultInfoSerializer,
		//transaction.View
		transaction_view.NewDefaultView,
		repository.NewTransactionRepository,
		service.NewCsvService,
		csv.NewRequestsMapper,

		handler.NewHandlerParams,
		handler.NewTransactionHandler,
		handler.NewCsvHandler,
		handler.NewHistoryHandler,
	}
}
