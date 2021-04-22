package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/jinzhu/gorm"
	"github.com/olebedev/emitter"

	"github.com/Confialink/wallet-accounts/internal/modules/request/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
)

// Canceller cancels pending request
type Canceller struct {
	db      *gorm.DB
	emitter *emitter.Emitter

	currencyProvider transfer.CurrencyProvider
	pf               transfers.PermissionFactory
}

func NewCanceller(
	db *gorm.DB,
	emitter *emitter.Emitter,
	currencyProvider transfer.CurrencyProvider,
	pf transfers.PermissionFactory,
) *Canceller {
	return &Canceller{db: db, emitter: emitter, currencyProvider: currencyProvider, pf: pf}
}

func (c *Canceller) Call(request *model.Request, reason string, currentUser *users.User) error {
	tx := c.db.Begin()

	canceller, err := transfers.CreateCanceller(tx, request, c.currencyProvider, c.pf)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = canceller.Cancel(request, reason)
	if err != nil {
		tx.Rollback()
		return err
	}

	eventContext := &event.ContextPendingRequestCancelled{
		Tx:        c.db,
		UserID:    *request.UserId,
		RequestID: *request.Id,
		Reason:    reason,
	}

	<-c.emitter.Emit(event.PendingRequestCancelled, eventContext)

	tx.Commit()

	return err
}
