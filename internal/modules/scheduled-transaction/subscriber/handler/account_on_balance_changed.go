package handler

import (
	accountEvent "github.com/Confialink/wallet-accounts/internal/modules/account/event"
	scheduledTransaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"
	"github.com/inconshreveable/log15"
	"github.com/olebedev/emitter"
	"github.com/shopspring/decimal"
)

func AccountOnBalanceChanged(
	eventEmitter *emitter.Emitter,
	scheduledTransactionService *scheduledTransaction.Service,
	logger log15.Logger,
) {
	logger = logger.New("eventHandler", "scheduled-transaction.AccountOnBalanceChanged")
	onBalanceChanged := func(event *emitter.Event) {
		context := event.Args[0].(*accountEvent.ContextAccountBalanceChanged)

		balanceLimitFee(context, scheduledTransactionService, logger)
		event.Flags = event.Flags | emitter.FlagSync
	}

	// must be used as middleware (synchronous call)
	// empty loop is aimed to free chanel once event is emitted
	for range eventEmitter.On(accountEvent.AccountBalanceChanged, onBalanceChanged) { /* empty */
	}

}

func balanceLimitFee(
	context *accountEvent.ContextAccountBalanceChanged,
	service *scheduledTransaction.Service,
	logger log15.Logger,
) {
	account := context.Account
	if account.Type.BalanceLimitAmount != nil && account.Type.BalanceFeeAmount != nil {
		if account.Balance.LessThan(*account.Type.BalanceLimitAmount) {

			feeAmount := *account.Type.BalanceFeeAmount
			if feeAmount.LessThanOrEqual(decimal.Zero) {
				logger.Warn("balance limit amount is specified however balance fee amount is incorrect", "feeAmount", feeAmount)
				return
			}

			chargeDay := 1
			if account.Type.BalanceChargeDay != nil {
				if *account.Type.BalanceChargeDay > 0 && *account.Type.BalanceChargeDay < 31 {
					chargeDay = int(*account.Type.BalanceChargeDay)
				}
			}

			err := service.ScheduleTransfer(
				&scheduledTransaction.ScheduleParams{
					Account:    context.Account,
					Reason:     scheduledTransaction.ReasonLimitBalanceFee,
					Period:     scheduledTransaction.PeriodMonthly,
					PaymentDay: chargeDay,
					Amount:     feeAmount.Neg(),
				},
				context.DbTransaction,
			)

			// ignore if already scheduled
			if err == scheduledTransaction.ErrorAlreadyScheduled {
				return
			}
		}
	}
}
