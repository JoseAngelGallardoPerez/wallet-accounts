package scheduled_transaction

import (
	"sync"

	"github.com/robfig/cron"

	"github.com/Confialink/wallet-pkg-utils"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/modules/request"
)

func Schedule(
	repo *Repository,
	requestCreator *request.Creator,
	db *gorm.DB,
	scheduler *Service,
	logger log15.Logger,
) (*cron.Cron, error) {
	mutex := sync.Mutex{}

	schedule, err := getScheduleConfig()
	if err != nil {
		return nil, err
	}

	timeSource := utils.NewTimeSource(schedule.Location)

	localizedCron := cron.NewWithLocation(schedule.Location)

	// Collect line of credit data
	localizedCron.Schedule(schedule.CollectCreditLineFee, cron.FuncJob(func() {
		logger.Info("running scheduled job: collect line of credit data")
		mutex.Lock()
		defer mutex.Unlock()
		WatchCreditLine(scheduler, db, logger, timeSource)
	}))
	// Collect account interest data
	localizedCron.Schedule(schedule.CollectAccountInterest, cron.FuncJob(func() {
		logger.Info("running scheduled job: collect account interest data")
		mutex.Lock()
		defer mutex.Unlock()
		WatchInterestGeneration(scheduler, db, logger, timeSource)
	}))
	// Collect maintenance fee
	localizedCron.Schedule(schedule.CollectMaintenanceFee, cron.FuncJob(func() {
		logger.Info("running scheduled job: collect maintenance fee")
		mutex.Lock()
		defer mutex.Unlock()
		WatchMaintenanceFee(scheduler, db, logger, timeSource)
	}))
	// Collect maintenance fee
	localizedCron.Schedule(schedule.CollectMinimumBalance, cron.FuncJob(func() {
		logger.Info("running scheduled job: collect minimum balance fee")
		mutex.Lock()
		defer mutex.Unlock()
		WatchLimitBalance(scheduler, db, logger, timeSource)
	}))
	// Charge account maintenance fee
	localizedCron.Schedule(schedule.ChargeAccountMaintenanceFee, cron.FuncJob(func() {
		logger.Info("running scheduled job: charge account maintenance fee")
		mutex.Lock()
		defer mutex.Unlock()
		ExecuteScheduledTransactions(
			getScheduledTransactions(ReasonMaintenanceFee, timeSource, repo, logger),
			repo,
			requestCreator,
			db,
			logger,
		)
	}))

	// Payout account interest
	localizedCron.Schedule(schedule.PayoutAccountInterest, cron.FuncJob(func() {
		logger.Info("running scheduled job: payout account interest")
		mutex.Lock()
		defer mutex.Unlock()
		ExecuteScheduledTransactions(
			getScheduledTransactions(ReasonInterestGeneration, timeSource, repo, logger),
			repo,
			requestCreator,
			db,
			logger,
		)
	}))

	// Charge minimum balance fee
	localizedCron.Schedule(schedule.ChargeMinimumBalance, cron.FuncJob(func() {
		logger.Info("running scheduled job: charge minimum balance fee")
		mutex.Lock()
		defer mutex.Unlock()
		ExecuteScheduledTransactions(
			getScheduledTransactions(ReasonLimitBalanceFee, timeSource, repo, logger),
			repo,
			requestCreator,
			db,
			logger,
		)
	}))

	// Charge line of credit fee
	localizedCron.Schedule(schedule.ChargeCreditLine, cron.FuncJob(func() {
		logger.Info("running scheduled job: charge line of credit fee")
		mutex.Lock()
		defer mutex.Unlock()
		ExecuteScheduledTransactions(
			getScheduledTransactions(ReasonCreditLineFee, timeSource, repo, logger),
			repo,
			requestCreator,
			db,
			logger,
		)
	}))

	return localizedCron, nil
}

func getScheduledTransactions(reason Reason, timeSource utils.Time, repo *Repository, logger log15.Logger) []*ScheduledTransaction {
	scheduledTransactions, err := repo.GetScheduled(string(reason), timeSource.Now())
	if err != nil {
		logger.Error("failed to retrieve scheduled transactions", "error", err)
		return make([]*ScheduledTransaction, 0)
	}
	return scheduledTransactions
}
