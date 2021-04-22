package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

const (
	LimitMaxTotalBalance                = "max_total_balance"
	LimitMaxTotalBalanceEnabled         = true
	LimitMaxTotalBalanceDefaultAmount   = .0
	LimitMaxTotalBalanceDefaultCurrency = "EUR"
)

type maxBalanceLimit struct {
	details            types.Details
	limitService       *limit.Service
	aggregationService *balance.AggregationService
	logger             log15.Logger
}

func NewMaxBalanceLimit(
	details types.Details,
	limitService *limit.Service,
	aggregationService *balance.AggregationService,
	logger log15.Logger,
) PermissionChecker {
	return &maxBalanceLimit{
		details:            details,
		limitService:       limitService,
		aggregationService: aggregationService,
		logger:             logger,
	}
}

func (m *maxBalanceLimit) Check() error {
	creditByUser := make(map[string]balance.AggregationResult)
	for _, detail := range m.details {
		if detail.Account != nil && detail.IsCredit() {
			uid := detail.Account.UserId
			result, ok := creditByUser[uid]
			if !ok {
				result = balance.AggregationResult{
					{ItemAmount: detail.Amount, ItemCurrencyCode: detail.CurrencyCode},
				}
				creditByUser[uid] = result
				continue
			}
			creditByUser[uid] = append(result, balance.AggregationItem{ItemAmount: detail.Amount, ItemCurrencyCode: detail.CurrencyCode})
		}
	}

	for uid, aggregation := range creditByUser {
		// find limit for the given user
		lim, err := m.limitService.FindOne(limit.Identifier{
			Name:     LimitMaxTotalBalance,
			Entity:   "user",
			EntityId: uid,
		})
		// if no limit found then no need to check
		if errors.Cause(err) == limit.ErrNotFound {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "failed to check max limit: limit service returned error")
		}
		if lim.Available().NoLimit() {
			continue
		}
		limitAmount := lim.Available().CurrencyAmount()
		totalBalance, err := m.aggregationService.GeneralTotalByUserId(uid, limitAmount.CurrencyCode())
		if err != nil {
			return errors.Wrap(err, "failed to check max limit: aggregation service returned error")
		}
		aggregation = append(aggregation, totalBalance)
		// this is sum of current total balance of all user accounts plus total transfer amount
		totalBalanceAfter, err := m.aggregationService.Reduce(aggregation, limitAmount.CurrencyCode())
		if err != nil {
			return errors.Wrap(err, "failed to check max limit: aggregation service failed to reduce total balance")
		}
		// finally check if the result fits into the limit
		if err = lim.WithinLimit(&totalBalanceAfter); err != nil {
			if errors.Cause(err) == limit.ErrLimitExceeded {
				// TODO: get rid of logger here, see https://velmie.atlassian.net/browse/VW-296
				err = errors.Wrapf(
					err,
					"max total balance is exceeded: user with id %s has limit %s %s, but the amount after the transfer would be %s %s",
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

func (m *maxBalanceLimit) Name() string {
	return LimitMaxTotalBalance
}
