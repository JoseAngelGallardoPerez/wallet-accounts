package handler

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	requestEvent "github.com/Confialink/wallet-accounts/internal/modules/request/event"
	txConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	pb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/olebedev/emitter"
)

func RequestOnPendingApproval(eventEmitter *emitter.Emitter) {
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(requestEvent.RequestPendingApproval, notifyPendingApprovalRequest) { /* empty */
	}
}

func notifyPendingApprovalRequest(event *emitter.Event) {
	context := event.Args[0].(*requestEvent.ContextRequestPending)
	request := context.Request

	if request.IsInitiatedByUser() {
		go notifyPendingRequestByUser(context)
		return
	}

	// OWT request by admin
	isOwtRequest := *request.Subject == constants.SubjectTransferOutgoingWireTransfer
	mustNotify := isOwtRequest && *request.IsInitiatedByAdmin

	if mustNotify {
		go notifyPendingOWTByAdmin(context)
	}
}

func notifyPendingRequestByUser(context *requestEvent.ContextRequestPending) {
	request := context.Request
	service := notificationService
	err := service.TriggerNewTransferRequest(*request.UserId)
	if err != nil {
		logger.Warn(
			"failed to trigger pending approval request notification",
			"error", err,
			"requestId", *request.Id,
		)
	}
}

func notifyPendingOWTByAdmin(context *requestEvent.ContextRequestPending) {
	request := context.Request
	details := context.Details
	outgoingDetail := details.ByPurpose(txConstants.PurposeOWTOutgoing)
	sourceAccount := outgoingDetail.Account

	service := notificationService
	users, err := usersService.GetByUIDs([]string{*request.UserId, sourceAccount.UserId})
	if err != nil {
		logger.Error("failed to retrieve users", "error", err)
		// do not set error to context since it is not critical error, we can skip event
		return
	}
	if len(users) != 2 {
		logger.Error("invalid count of users")
		// do not set error to context since it is not critical error, we can skip event
		return
	}

	var admin, client *pb.User
	for _, usr := range users {
		if usr.GetUID() == *request.UserId {
			admin = usr
		} else {
			client = usr
		}
	}

	err = service.TriggerNewTransferRequestByAdmin(client, admin)
	if err != nil {
		logger.Warn(
			"failed to trigger OWT pending approval request notification",
			"error", err,
			"requestId", *request.Id,
		)
	}
}
