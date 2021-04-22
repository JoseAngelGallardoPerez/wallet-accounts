package subscriber

import (
	"log"

	"github.com/Confialink/wallet-accounts/internal/modules/notifications/subscriber/handler"
	"github.com/olebedev/emitter"
)

func Subscribe(eventEmitter *emitter.Emitter) {
	go handler.RequestOnPendingApproval(eventEmitter)
	go handler.RequestOnRequestExecuted(eventEmitter)
	go handler.RequestOnRequestCancelled(eventEmitter)
	log.Println("module notifications subscribed on application events")
}
