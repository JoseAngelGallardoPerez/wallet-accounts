package scheduled_transaction_provider

import (
	scheduled_transaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"
	"github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction/http/handler"
)

func Providers() []interface{} {
	return []interface{}{
		scheduled_transaction.NewRepository,
		scheduled_transaction.NewScheduledTransactionLogRepository,
		scheduled_transaction.NewService,

		handler.NewTransactionsHandler,
	}
}
