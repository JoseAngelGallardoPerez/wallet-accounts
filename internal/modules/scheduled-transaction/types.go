package scheduled_transaction

import (
	"errors"
	"strings"
	"time"

	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/shopspring/decimal"
)

type Reason string

const (
	ReasonMaintenanceFee     = Reason("maintenance_fee")
	ReasonLimitBalanceFee    = Reason("limit_balance_fee")
	ReasonCreditLineFee      = Reason("credit_line_fee")
	ReasonInterestGeneration = Reason("interest_generation")
)

var reasonHumanReadable = map[Reason]string{
	ReasonMaintenanceFee:     "Monthly Maintenance Fee",
	ReasonLimitBalanceFee:    "Minimum CurrentBalance Fee",
	ReasonInterestGeneration: "Account Interest Payout",
	ReasonCreditLineFee:      "Line of CreditFromAlias Fee",
}

func (r Reason) Description() string {
	if description, set := reasonHumanReadable[r]; set {
		return description
	}
	return strings.Replace(string(r), "_", " ", -1)
}

type Status string

const (
	StatusPending  = Status("pending")
	StatusExecuted = Status("executed")
)

type Period string

const (
	PeriodMonthly    = "Monthly"
	PeriodQuarterly  = "Quarterly"
	PeriodBiAnnually = "Bi-Annually"
	PeriodAnnually   = "Annually"
)

var knownPeriods = map[string]Period{
	string(PeriodMonthly):    PeriodMonthly,
	string(PeriodQuarterly):  PeriodQuarterly,
	string(PeriodBiAnnually): PeriodBiAnnually,
	string(PeriodAnnually):   PeriodAnnually,
}

func PeriodFromString(period string) (Period, error) {
	if result, ok := knownPeriods[period]; ok {
		return result, nil
	}
	return Period(""), errors.New("unknown period " + period)
}

type ScheduleParams struct {
	Account    *model.Account
	Reason     Reason
	Amount     decimal.Decimal
	Period     Period
	PaymentDay int
	// Month overrides found month if specified
	Month *time.Month
	Now   time.Time // current date can be changed when we run simulations
}
