package subscriber

import (
	"log"

	scheduledTransaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"
	"github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction/subscriber/handler"
	"github.com/inconshreveable/log15"
	"github.com/olebedev/emitter"
)

func Subscribe(
	eventEmitter *emitter.Emitter,
	scheduledTransactionService *scheduledTransaction.Service,
	logger log15.Logger,
) {
	go handler.AccountOnBalanceChanged(eventEmitter, scheduledTransactionService, logger)
	log.Println("module scheduled-transaction subscribed on application events")
}
