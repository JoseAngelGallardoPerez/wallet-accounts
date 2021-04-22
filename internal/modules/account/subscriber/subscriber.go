package subscriber

import (
	"log"

	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account/subscriber/handler"
	"github.com/inconshreveable/log15"
	"github.com/olebedev/emitter"
)

func Subscribe(
	eventEmitter *emitter.Emitter,
	accountsRepository *repository.AccountRepository,
	logger log15.Logger,
) {
	go handler.AccountTypeOnUpdate(eventEmitter, accountsRepository, logger)
	log.Println("module account subscribed on application events")
}
