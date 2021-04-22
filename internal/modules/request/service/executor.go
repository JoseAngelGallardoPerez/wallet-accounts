package service

import (
	accountEvent "github.com/Confialink/wallet-accounts/internal/modules/account/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/jinzhu/gorm"
	"github.com/olebedev/emitter"

	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
)

// Executor executes pending request with needed after actions
type Executor struct {
	db      *gorm.DB
	emitter *emitter.Emitter

	currencyProvider transfer.CurrencyProvider
	pf               transfers.PermissionFactory
}

func NewExecutor(
	db *gorm.DB,
	emitter *emitter.Emitter,
	currencyProvider transfer.CurrencyProvider,
	pf transfers.PermissionFactory,
) *Executor {
	return &Executor{db, emitter, currencyProvider, pf}
}

func (e *Executor) Call(request *model.Request, currentUser *users.User) error {
	tx := e.db.Begin()

	executor, err := transfers.CreateExecutor(tx, request, e.currencyProvider, e.pf)

	if err != nil {
		tx.Rollback()
		return err
	}

	details, err := executor.Execute(request)
	if err != nil {
		tx.Rollback()
		return err
	}

	eventContext := &event.ContextRequestExecuted{
		Tx:      tx,
		Request: request,
		Details: details,
	}

	<-e.emitter.Emit(event.RequestExecuted, eventContext)
	accountEvent.TriggerBalanceChanged(e.emitter, tx, *request.Subject, details)
	tx.Commit()

	return err
}
