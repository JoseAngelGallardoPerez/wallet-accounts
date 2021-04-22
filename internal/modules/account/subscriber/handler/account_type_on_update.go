package handler

import (
	accountTypeEvent "github.com/Confialink/wallet-accounts/internal/modules/account-type/event"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/inconshreveable/log15"
	"github.com/olebedev/emitter"
	"github.com/shopspring/decimal"
)

func AccountTypeOnUpdate(
	eventEmitter *emitter.Emitter,
	accountsRepository *repository.AccountRepository,
	logger log15.Logger,
) {
	logger = logger.New("eventHandler", "account.AccountTypeOnUpdate")
	onUpdate := func(event *emitter.Event) {
		context := event.Args[0].(*accountTypeEvent.ContextAccountTypeUpdated)

		oldData := context.OldAccountType
		newData := context.NewAccountType

		// Ensure values are not nil pointers
		if oldData.CreditLimitAmount == nil {
			oldData.CreditLimitAmount = &decimal.Zero
		}
		if newData.CreditLimitAmount == nil {
			newData.CreditLimitAmount = &decimal.Zero
		}
		// Calculate difference between old and new values
		diff := newData.CreditLimitAmount.Sub(*oldData.CreditLimitAmount)
		// if difference exists then we must update all related accounts
		if !diff.Equal(decimal.Zero) {
			accountsRepository := accountsRepository.WrapContext(context.DbTransaction)
			context.Error = accountsRepository.UpdateAvailableAmountByAccountTypeId(diff, newData.ID)
			if context.Error != nil {
				logger.Error("failed to update available balance value", "error", context.Error, "accountTypeId", newData.ID, "diff", diff)
			}
		}
	}

	// on update must be used as middleware (synchronous call)
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(accountTypeEvent.AccountTypeUpdated, onUpdate) { /* empty */
	}
}
