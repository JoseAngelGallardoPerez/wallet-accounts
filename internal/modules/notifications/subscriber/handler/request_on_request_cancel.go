package handler

import (
	"github.com/inconshreveable/log15"
	"github.com/olebedev/emitter"

	"github.com/Confialink/wallet-accounts/internal/modules/notifications"
	requestEvent "github.com/Confialink/wallet-accounts/internal/modules/request/event"
)

func RequestOnRequestCancelled(eventEmitter *emitter.Emitter) {
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(requestEvent.PendingRequestCancelled, notifyRequestCancelled) { /* empty */
	}
}

func notifyRequestCancelled(event *emitter.Event) {
	logger := logger.New("eventHandler", "notification.notifyRequestCancelled")
	context := event.Args[0].(*requestEvent.ContextPendingRequestCancelled)

	go processNotifyCancelled(context, notificationService, logger)
}

func processNotifyCancelled(context *requestEvent.ContextPendingRequestCancelled, service *notifications.Service, logger log15.Logger) {
	if err := service.TriggerRequestCancelled(context.UserID, context.RequestID); err != nil {
		logger.Error("failed to notify", "error", err, "userID", context.UserID, "requestID", context.RequestID)
	}
}
