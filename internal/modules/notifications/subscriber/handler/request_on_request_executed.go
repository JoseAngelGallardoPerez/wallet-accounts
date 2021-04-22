package handler

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	txConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/inconshreveable/log15"
	"github.com/olebedev/emitter"

	"github.com/Confialink/wallet-accounts/internal/modules/notifications"
	requestEvent "github.com/Confialink/wallet-accounts/internal/modules/request/event"
)

func RequestOnRequestExecuted(eventEmitter *emitter.Emitter) {
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(requestEvent.RequestExecuted, notifyRequestExecuted) { /* empty */
	}
}

func notifyRequestExecuted(event *emitter.Event) {
	logger := logger.New("eventHandler", "notification.notifyRequestExecuted")
	context := event.Args[0].(*requestEvent.ContextRequestExecuted)

	go processNotify(context, notificationService, logger)
}

func processNotify(context *requestEvent.ContextRequestExecuted, service *notifications.Service, logger log15.Logger) {
	if err := service.TriggerRequestExecuted(context.Request); err != nil {
		logger.Error("failed to notify", "error", err, "userID", *context.Request.UserId, "requestID", *context.Request.Id)
	}

	request := context.Request
	switch *request.Subject {
	case constants.SubjectTransferBetweenUsers:
		notifyTBU(context, service, logger)
	case constants.SubjectTransferBetweenAccounts:
		notifyTBA(context, service, logger)
	case constants.SubjectCreditAccount:
		notifyCA(context, service, logger)
	}
}

func notifyCA(
	context *requestEvent.ContextRequestExecuted,
	service *notifications.Service,
	logger log15.Logger,
) {
	request := context.Request
	if request.GetInput().GetBool("isInitialBalanceRequest") {
		return
	}
	details := context.Details
	methods := []string{
		"email",
		"push_notification",
	}
	incomingDetail := details.ByPurpose(txConstants.PurposeCreditAccount)
	destinationAccount := incomingDetail.Account
	err := service.TriggerIncomingTransaction(destinationAccount.UserId, destinationAccount, methods)

	if err != nil {
		logger.Warn(
			"failed to trigger CA incoming transaction notification",
			"error", err,
			"requestId", *request.Id,
		)
	}
}

func notifyTBA(
	context *requestEvent.ContextRequestExecuted,
	service *notifications.Service,
	logger log15.Logger,
) {
	request := context.Request
	if *request.IsInitiatedBySystem || *request.IsInitiatedByAdmin {
		return
	}
	actionRequired, err := settingsService.Bool("tba_action_required")
	if err != nil {
		logger.Warn(
			"failed to check if TBA action is required",
			"error", err,
			"requestId", *request.Id,
		)
	}
	// if user initiated transfer and no admin approval required, then we don't need to notify user
	if !actionRequired {
		return
	}
	details := context.Details

	methods := []string{
		"email",
		"push_notification",
	}
	incomingDetail := details.ByPurpose(txConstants.PurposeTBAIncoming)
	destinationAccount := incomingDetail.Account
	err = service.TriggerIncomingTransaction(destinationAccount.UserId, destinationAccount, methods)

	if err != nil {
		logger.Warn(
			"failed to trigger TBA incoming transaction notification",
			"error", err,
			"requestId", *request.Id,
		)
	}

}

func notifyTBU(
	context *requestEvent.ContextRequestExecuted,
	service *notifications.Service,
	logger log15.Logger,
) {
	request := context.Request
	details := context.Details

	outgoingDetail := details.ByPurpose(txConstants.PurposeTBUOutgoing)
	err := service.TriggerOutgoingTransaction(
		outgoingDetail.Account.UserId,
		outgoingDetail.Transaction,
	)
	if err != nil {
		logger.Warn(
			"failed to trigger TBU outgoing transaction notification",
			"error", err,
			"requestId", *request.Id,
		)
	}

	if *request.IsInitiatedByAdmin {
		return
	}

	methods := []string{
		"email",
		"push_notification",
		"internal_message",
	}
	incomingDetail := details.ByPurpose(txConstants.PurposeTBUIncoming)
	destinationAccount := incomingDetail.Account
	err = service.TriggerIncomingTransaction(destinationAccount.UserId, destinationAccount, methods)

	if err != nil {
		logger.Warn(
			"failed to trigger TBU incoming transaction notification",
			"error", err,
			"requestId", *request.Id,
		)
	}
}
