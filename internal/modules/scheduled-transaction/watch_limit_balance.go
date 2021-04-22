package scheduled_transaction

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-pkg-utils"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const watchLimitBalanceMaxErrors = 3

func WatchLimitBalance(scheduler *Service, db *gorm.DB, logger log15.Logger, timeSource utils.Time) {
	logger = logger.New("Task", "WatchLimitBalance")
	getPaymentDay := func(account *accountModel.Account) int {
		day := 1
		if account.Type.BalanceChargeDay != nil {
			if *account.Type.BalanceChargeDay > 0 && *account.Type.BalanceChargeDay <= 31 {
				day = int(*account.Type.BalanceChargeDay)
			}
		}
		return day
	}

	errorsCount := 0
	successfullyScheduledTransfersCount := 0
	now := timeSource.Now()
	for _, account := range findAccountsHavingBalanceLessThanLimitBalance(db, timeSource) {
		err := scheduler.ScheduleTransfer(
			&ScheduleParams{
				Amount:     account.Type.BalanceFeeAmount.Neg(),
				PaymentDay: getPaymentDay(account),
				Period:     PeriodMonthly,
				Reason:     ReasonLimitBalanceFee,
				Account:    account,
				Now:        now,
			},
			db,
		)

		if err != nil {
			errorsCount++
			logger.Error("failed to schedule transfer", "error", err)
			if errorsCount >= watchLimitBalanceMaxErrors {
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
			ReasonLimitBalanceFee,
		)
	}
}

// findAccountsHavingBalanceLessThanLimitBalance searches for accounts
// having balance less than balance fee limit which are not scheduled for the fee charging
func findAccountsHavingBalanceLessThanLimitBalance(db *gorm.DB, timeSource utils.Time) []*accountModel.Account {
	var accounts []*accountModel.Account

	today := timeSource.Now()
	db.
		Table("accounts").
		Preload("Type").
		Select("distinct accounts.*").
		Joins("inner join account_types act on act.id = accounts.type_id").
		Joins("left join scheduled_transactions st ON st.account_id = accounts.id").
		Where(`act.balance_limit_amount > 0 
					 and act.balance_fee_amount > 0 
					 and accounts.balance < act.balance_limit_amount
					 and (accounts.maturity_date > ? OR accounts.maturity_date IS NULL)
					 and (st.reason is null or st.reason = ?)
					 and (st.status is null or st.status != ?)
					 and accounts.id not in 
						(select account_id from scheduled_transactions where reason = ? and status = ?)`,
			today,
			ReasonLimitBalanceFee,
			StatusPending,
			ReasonLimitBalanceFee,
			StatusPending,
		).
		Find(&accounts)

	return accounts
}
