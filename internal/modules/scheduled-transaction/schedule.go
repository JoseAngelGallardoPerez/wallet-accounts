package scheduled_transaction

import (
	"time"

	"github.com/robfig/cron"
)

var getScheduleConfig ScheduleConfigurator = DefaultSchedule

type ScheduleConfig struct {
	Location               *time.Location
	CollectCreditLineFee   cron.Schedule
	CollectAccountInterest cron.Schedule
	CollectMaintenanceFee  cron.Schedule
	CollectMinimumBalance  cron.Schedule

	ChargeCreditLine            cron.Schedule
	PayoutAccountInterest       cron.Schedule
	ChargeMinimumBalance        cron.Schedule
	ChargeAccountMaintenanceFee cron.Schedule
}

type ScheduleConfigurator func() (*ScheduleConfig, error)

func DefaultSchedule() (*ScheduleConfig, error) {
	// Second | Minute | Hour | Dom(day of month) | Month | DowOptional(day of week optional) | Descriptor
	collectCreditLine, _ := cron.Parse("0 10 0 * *")       // at 00:10 every day
	collectAccountInterest, _ := cron.Parse("0 30 0 * *")  // at 00:30 every day
	collectMaintenanceFee, _ := cron.Parse("0 */10 * * *") // every 10 minutes
	collectMinimumBalance, _ := cron.Parse("0 */12 * * *") // every 12 minutes

	chargeCreditLine, _ := cron.Parse("0 20 0 * *")             // at 00:20 every day
	payoutAccountInterest, _ := cron.Parse("0 40 0 * *")        // at 00:40 every day
	chargeMinimumBalance, _ := cron.Parse("0 50 0 * *")         // at 00:50 every day
	chargeAccountMaintenanceFee, _ := cron.Parse("0 0 1 * * *") // at 01:00 every day

	return &ScheduleConfig{
		Location:               time.FixedZone("UTC+1(CET)", 3600),
		CollectCreditLineFee:   collectCreditLine,
		CollectAccountInterest: collectAccountInterest,
		CollectMaintenanceFee:  collectMaintenanceFee,
		CollectMinimumBalance:  collectMinimumBalance,

		ChargeCreditLine:            chargeCreditLine,
		PayoutAccountInterest:       payoutAccountInterest,
		ChargeMinimumBalance:        chargeMinimumBalance,
		ChargeAccountMaintenanceFee: chargeAccountMaintenanceFee,
	}, nil
}

func SetScheduleConfigurator(configurator ScheduleConfigurator) {
	getScheduleConfig = configurator
}
