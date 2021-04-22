package handler

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	requestEvent "github.com/Confialink/wallet-accounts/internal/modules/request/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/jinzhu/gorm"
	"github.com/olebedev/emitter"
)

func RequestOnRequestModified(eventEmitter *emitter.Emitter) {
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(requestEvent.RequestModified, handleRequestModified) { /* empty */
	}
}

func RequestOnPendingApproval(eventEmitter *emitter.Emitter) {
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(requestEvent.RequestPendingApproval, handleRequestPending) { /* empty */
	}
}

func RequestOnRequestExecuted(eventEmitter *emitter.Emitter) {
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(requestEvent.RequestExecuted, handleRequestExecuted) { /* empty */
	}
}

func handleRequestModified(event *emitter.Event) {
	context := event.Args[0].(*requestEvent.ContextRequestModified)
	makeSnapshot(context.Request, context.Details, context.Tx)
}

func handleRequestPending(event *emitter.Event) {
	context := event.Args[0].(*requestEvent.ContextRequestPending)
	makeSnapshot(context.Request, context.Details, context.Tx)
}

func handleRequestExecuted(event *emitter.Event) {
	context := event.Args[0].(*requestEvent.ContextRequestExecuted)
	makeSnapshot(context.Request, context.Details, context.Tx)
}

func makeSnapshot(request *model.Request, details types.Details, db *gorm.DB) {
	srv := snapshotService
	if db != nil {
		srv = srv.WrapContext(db)
	}
	done := make(map[balance.Balance]bool)
	for _, detail := range details {
		b := detailToBalance(detail)
		if _, alreadyDone := done[b]; alreadyDone {
			continue
		}
		done[b] = true
		_, err := srv.MakeSnapshot(request, b)
		if err != nil {
			logger.Error(
				"failed to create balance snapshot",
				"error", err,
				"requestId", request.Id,
				"detail", detail.GoString(),
			)
		}
	}
}

func detailToBalance(detail *types.Detail) balance.Balance {
	if detail.Account != nil {
		return detail.Account
	}
	if detail.Card != nil {
		return detail.Card
	}
	if detail.RevenueAccount != nil {
		return detail.RevenueAccount
	}
	panic("expected that detail contains account, card or revenue account")
}
