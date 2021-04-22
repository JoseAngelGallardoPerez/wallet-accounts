package scheduled_transaction

import (
	"time"

	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/calculation"
	"github.com/Confialink/wallet-pkg-utils"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const watchInterestGenerationMaxErrors = 3

func WatchInterestGeneration(scheduler *Service, db *gorm.DB, logger log15.Logger, timeSource utils.Time) {
	logger = logger.New("Task", "WatchInterestGeneration")

	getPaymentDay := func(account *accountModel.Account) int {
		if account.PayoutDay != nil {
			if *account.PayoutDay > 0 && *account.PayoutDay <= 31 {
				return int(*account.PayoutDay)
			}
		}
		if account.Type.DepositPayoutDay != nil {
			if *account.Type.DepositPayoutDay > 0 && *account.Type.DepositPayoutDay < 31 {
				return int(*account.Type.DepositPayoutDay)
			}
		}
		return 1
	}

	errorsCount := 0
	successfullyScheduledTransfersCount := 0
	now := timeSource.Now()
	for _, account := range findAccountsForDailyInterestGeneration(db, timeSource) {
		var period Period
		method, err := calculation.MethodFromString(account.Type.DepositPayoutMethod.Method)
		if err == nil {
			period, err = PeriodFromString(account.Type.DepositPayoutPeriod.Name)
		}

		if err == nil {
			feePercent := *account.Type.DepositAnnualInterestRate

			params := &ScheduleParams{
				Amount:     method.AnnualInterest(account.Balance, feePercent),
				PaymentDay: getPaymentDay(account),
				Period:     period,
				Reason:     ReasonInterestGeneration,
				Account:    account,
				Now:        now,
			}
			if account.Type.DepositPayoutMonth != nil {
				m := time.Month(*account.Type.DepositPayoutMonth)
				params.Month = &m
			}
			err = scheduler.ScheduleTransfer(params, db)
		}

		if err != nil {
			errorsCount++
			logger.Error("failed to schedule transfer", "error", err)
			if errorsCount >= watchInterestGenerationMaxErrors {
				break
			}
			continue
		}
		successfullyScheduledTransfersCount++
	}

	if successfullyScheduledTransfersCount > 0 {
		logger.Info(
			"successfully scheduled transfers",
			"count",
			successfullyScheduledTransfersCount,
			"reason",
			ReasonInterestGeneration,
		)
	}
}

func findAccountsForDailyInterestGeneration(db *gorm.DB, timeSource utils.Time) []*accountModel.Account {
	var accounts []*accountModel.Account

	today := timeSource.Now()
	db.
		Table("accounts").
		Preload("Type").
		Preload("Type.DepositPayoutMethod").
		Preload("Type.DepositPayoutPeriod").
		Select("distinct accounts.*").
		Joins("inner join account_types act on act.id = accounts.type_id").
		Joins("inner join payout_methods pm on pm.id = act.deposit_payout_method_id").
		Where(`act.deposit_annual_interest_rate > 0 
					 and act.deposit_payout_method_id is not null
					 and act.deposit_payout_period_id is not null
					 and act.deposit_payout_day is not null
				     and accounts.balance > 0
					 and (accounts.maturity_date > ? OR accounts.maturity_date IS NULL)
					 and pm.method = ?
					 and accounts.id not in (
						select st.account_id from scheduled_transactions st
						inner join scheduled_transaction_logs logs on logs.scheduled_transaction_id = st.id
						where st.reason = ?
						and logs.created_at between ? and ?)`,
			today,
			calculation.InterestCalculationMethodDaily,
			ReasonInterestGeneration,
			timeSource.BeginningOfDay(),
			timeSource.EndOfDay()).
		Find(&accounts)

	return accounts
}
