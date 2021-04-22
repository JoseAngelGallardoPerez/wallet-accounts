package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"time"
)

const (
	LimitMaxTotalDebitPerDay                = "max_total_debit_per_day"
	LimitMaxTotalDebitPerDayEnabled         = true
	LimitMaxTotalDebitPerDayDefaultAmount   = .0
	LimitMaxTotalDebitPerDayDefaultCurrency = "EUR"

	LimitMaxTotalDebitPerMonth                = "max_total_debit_per_month"
	LimitMaxTotalDebitPerMonthEnabled         = true
	LimitMaxTotalDebitPerMonthDefaultAmount   = .0
	LimitMaxTotalDebitPerMonthDefaultCurrency = "EUR"
)

type maxTotalDebitPerPeriod struct {
	details            types.Details
	limitService       *limit.Service
	aggregationService *balance.AggregationService
	limitName          string
	periodFrom         time.Time
	periodTill         time.Time
	logger             log15.Logger
}

func NewMaxTotalDebitPerPeriod(
	details types.Details,
	limitService *limit.Service,
	aggregationService *balance.AggregationService,
	limitName string,
	periodFrom time.Time,
	periodTill time.Time,
	logger log15.Logger,
) PermissionChecker {
	return &maxTotalDebitPerPeriod{
		details:            details,
		limitService:       limitService,
		aggregationService: aggregationService,
		limitName:          limitName,
		periodFrom:         periodFrom,
		periodTill:         periodTill,
		logger:             logger,
	}
}

func (m *maxTotalDebitPerPeriod) Check() error {
	debitByUser := make(map[string]balance.AggregationResult)
	for _, detail := range m.details {
		if detail.Account != nil && detail.IsDebit() {
			uid := detail.Account.UserId
			result, ok := debitByUser[uid]
			if !ok {
				result = balance.AggregationResult{
					{ItemAmount: detail.Amount.Abs(), ItemCurrencyCode: detail.CurrencyCode},
				}
				debitByUser[uid] = result
				continue
			}
			debitByUser[uid] = append(result, balance.AggregationItem{ItemAmount: detail.Amount.Abs(), ItemCurrencyCode: detail.CurrencyCode})
		}
	}

	for uid, aggregation := range debitByUser {
		// find limit for the given user
		lim, err := m.limitService.FindOne(limit.Identifier{
			Name:     m.limitName,
			Entity:   "user",
			EntityId: uid,
		})
		// if no limit found then no need to check
		if errors.Cause(err) == limit.ErrNotFound {
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "failed to check %s: limit service returned error", m.limitName)
		}
		if lim.Available().NoLimit() {
			continue
		}
		limitAmount := lim.Available().CurrencyAmount()
		totalBalance, err := m.aggregationService.TotalDebitedByUserPerPeriod(uid, m.periodFrom, m.periodTill, limitAmount.CurrencyCode())
		if err != nil {
			return errors.Wrapf(err, "failed to check %s: aggregation service returned error", m.limitName)
		}
		aggregation = append(aggregation, totalBalance)
		// this is sum of current total balance of all user accounts plus total transfer amount
		totalBalanceAfter, err := m.aggregationService.Reduce(aggregation, limitAmount.CurrencyCode())
		if err != nil {
			return errors.Wrapf(err, "failed to check %s: aggregation service failed to reduce total balance", m.limitName)
		}
		// finally check if the result fits into the limit
		if err = lim.WithinLimit(&totalBalanceAfter); err != nil {
			if errors.Cause(err) == limit.ErrLimitExceeded {
				// TODO: get rid of logger here, see https://velmie.atlassian.net/browse/VW-296
				err = errors.Wrapf(
					err,
					"%s is exceeded: user with id %s has limit %s %s, but the amount after the transfer would be %s %s",
					m.limitName,
					uid,
					limitAmount.Amount().String(),
					limitAmount.CurrencyCode(),
					totalBalanceAfter.ItemAmount.String(),
					totalBalanceAfter.ItemCurrencyCode,
				)
				m.logger.Info(err.Error())
				return err
			}
			return err
		}
	}
	return nil
}

func (m *maxTotalDebitPerPeriod) Name() string {
	return m.limitName
}
