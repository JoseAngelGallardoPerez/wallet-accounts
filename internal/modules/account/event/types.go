package event

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/jinzhu/gorm"
	"github.com/olebedev/emitter"
)

const (
	AccountBalanceChanged = "account:balance-changed"
)

type ContextAccountBalanceChanged struct {
	// database transaction context
	DbTransaction *gorm.DB
	// related account
	Account *model.Account
	// Subject of request
	RequestSubject constants.Subject
}

func TriggerBalanceChanged(eventEmitter *emitter.Emitter, tx *gorm.DB, subject constants.Subject, details types.Details) {
	processed := make(map[uint64]struct{})
	for _, detail := range details {
		if detail.Account != nil {
			account := detail.Account
			if _, skip := processed[account.ID]; skip {
				continue
			}
			eventContext := &ContextAccountBalanceChanged{
				DbTransaction:  tx,
				Account:        account,
				RequestSubject: subject,
			}
			<-eventEmitter.Emit(AccountBalanceChanged, eventContext)
			processed[account.ID] = struct{}{}
		}
	}
}
