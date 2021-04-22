package subscriber

import (
	"log"

	"github.com/Confialink/wallet-accounts/internal/modules/balance/subscriber/handler"
	"github.com/olebedev/emitter"
)

func Subscribe(eventEmitter *emitter.Emitter) {
	go handler.RequestOnPendingApproval(eventEmitter)
	go handler.RequestOnRequestExecuted(eventEmitter)
	go handler.RequestOnRequestModified(eventEmitter)
	log.Println("module balance subscribed on application events")
}
