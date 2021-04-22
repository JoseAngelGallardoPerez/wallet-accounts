package scheduled_transaction

import (
	"time"

	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/calculation"
	"github.com/Confialink/wallet-pkg-utils"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const watchCreditLineMaxErrors = 3

func WatchCreditLine(scheduler *Service, db *gorm.DB, logger log15.Logger, timeSource utils.Time) {
	logger = logger.New("Task", "WatchCreditLine")

	getPaymentDay := func(account *accountModel.Account) int {
		day := 1
		if account.Type.CreditChargeDay != nil {
			if *account.Type.CreditChargeDay > 0 && *account.Type.CreditChargeDay <= 31 {
				day = int(*account.Type.CreditChargeDay)
			}
		}
		return day
	}

	errorsCount := 0
	successfullyScheduledTransfersCount := 0
	now := timeSource.Now()
	for _, account := range findAccountsForDailyCreditFee(db, timeSource) {
		var period Period
		method, err := calculation.MethodFromString(account.Type.CreditPayoutMethod.Method)
		if err == nil {
			period, err = PeriodFromString(account.Type.CreditChargePeriod.Name)
		}

		if err == nil {
			feePercent := *account.Type.CreditAnnualInterestRate

			params := &ScheduleParams{
				Amount:     method.AnnualInterest(account.Balance.Abs(), feePercent).Neg(),
				PaymentDay: getPaymentDay(account),
				Period:     period,
				Reason:     ReasonCreditLineFee,
				Account:    account,
				Now:        now,
			}

			if account.Type.CreditChargeMonth != nil {
				m := time.Month(*account.Type.CreditChargeMonth)
				params.Month = &m
			}

			err = scheduler.ScheduleTransfer(params, db)
		}

		if err != nil {
			errorsCount++
			logger.Error("failed to schedule transfer", "error", err)
			if errorsCount >= watchCreditLineMaxErrors {
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
			ReasonCreditLineFee,
		)
	}
}

func findAccountsForDailyCreditFee(db *gorm.DB, timeSource utils.Time) []*accountModel.Account {
	var accounts []*accountModel.Account

	today := timeSource.Now()
	db.
		Table("accounts").
		Preload("Type").
		Preload("Type.CreditPayoutMethod").
		Preload("Type.CreditChargePeriod").
		Select("distinct accounts.*").
		Joins("inner join account_types act on act.id = accounts.type_id").
		Joins("inner join payout_methods pm on pm.id = act.credit_payout_method_id").
		Where(`act.credit_limit_amount > 0 
					 and act.credit_annual_interest_rate > 0 
				     and accounts.balance < 0
					 and (accounts.maturity_date > ? OR accounts.maturity_date IS NULL)
					 and pm.method = ?
					 and accounts.id not in (
						select st.account_id from scheduled_transactions st
						inner join scheduled_transaction_logs logs on logs.scheduled_transaction_id = st.id
						where st.reason = ?
						and logs.created_at between ? and ?)`,
			today,
			calculation.InterestCalculationMethodDaily,
			ReasonCreditLineFee,
			timeSource.BeginningOfDay(),
			timeSource.EndOfDay()).
		Find(&accounts)

	return accounts
}
