package scheduled_transaction

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-pkg-utils"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const watchLimitMaintenanceMaxErrors = 3

func WatchMaintenanceFee(scheduler *Service, db *gorm.DB, logger log15.Logger, timeSource utils.Time) {
	logger = logger.New("Task", "WatchMonthlyMaintenance")

	errorsCount := 0
	successfullyScheduledTransfersCount := 0
	now := timeSource.Now()
	for _, account := range findAccountsForMonthlyMaintenance(db, timeSource) {
		err := scheduler.ScheduleTransfer(
			&ScheduleParams{
				Amount:     account.Type.MonthlyMaintenanceFee.Neg(),
				PaymentDay: 1,
				Period:     PeriodMonthly,
				Reason:     ReasonMaintenanceFee,
				Account:    account,
				Now:        now,
			},
			db,
		)

		if err != nil {
			errorsCount++
			logger.Error("failed to schedule transfer", "error", err)
			if errorsCount >= watchLimitMaintenanceMaxErrors {
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
			ReasonMaintenanceFee,
		)
	}
}

// findAccountsForMonthlyMaintenance searches for accounts
// which have to be scheduled for monthly maintenance fee
func findAccountsForMonthlyMaintenance(db *gorm.DB, timeSource utils.Time) []*accountModel.Account {
	var accounts []*accountModel.Account

	today := timeSource.Now()
	db.
		Table("accounts").
		Preload("Type").
		Select("distinct accounts.*").
		Joins("inner join account_types act on act.id = accounts.type_id").
		Joins("left join scheduled_transactions st ON st.account_id = accounts.id").
		Where(`act.monthly_maintenance_fee > 0
					 and (accounts.maturity_date > ? OR accounts.maturity_date IS NULL)
					 and (st.reason is null or st.reason = ?)
					 and (st.status is null or st.status != ?)					 
					 and accounts.id not in 
						(select account_id from scheduled_transactions where reason = ? and status = ?)`,
			today,
			ReasonMaintenanceFee,
			StatusPending,
			ReasonMaintenanceFee,
			StatusPending,
		).
		Find(&accounts)

	return accounts
}
