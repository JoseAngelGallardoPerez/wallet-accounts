package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

const (
	// LimitMaxCreditPerTransfer is responsible for limiting the amount that a user
	// can receive(be credited) within single transfer
	LimitMaxCreditPerTransfer                = "max_credit_per_transfer"
	LimitMaxCreditPerTransferEnabled         = true
	LimitMaxCreditPerTransferDefaultAmount   = .0
	LimitMaxCreditPerTransferDefaultCurrency = "EUR"
)

type maxCreditPerTransfer struct {
	details      types.Details
	limitService *limit.Service
	reducer      balance.Reducer
	logger       log15.Logger
}

func NewMaxCreditPerTransfer(
	details types.Details,
	limitService *limit.Service,
	reducer balance.Reducer,
	logger log15.Logger,
) PermissionChecker {
	return &maxCreditPerTransfer{
		details:      details,
		limitService: limitService,
		reducer:      reducer,
		logger:       logger,
	}
}

func (m *maxCreditPerTransfer) Check() error {
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
			Name:     LimitMaxCreditPerTransfer,
			Entity:   "user",
			EntityId: uid,
		})
		// if no limit found then no need to check
		if errors.Cause(err) == limit.ErrNotFound {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "failed to check max credit per transfer: limit service returned error")
		}
		if lim.Available().NoLimit() {
			continue
		}
		limitAmount := lim.Available().CurrencyAmount()
		totalCredit, err := m.reducer.Reduce(aggregation, limitAmount.CurrencyCode())
		if err != nil {
			return errors.Wrap(err, "failed to check max credit per transfer: balance reducer returned error")
		}

		if err = lim.WithinLimit(&totalCredit); err != nil {
			if errors.Cause(err) == limit.ErrLimitExceeded {
				// TODO: get rid of logger here, see https://velmie.atlassian.net/browse/VW-296
				err = errors.Wrapf(
					err,
					"max credit per transfer limit is exceeded: user with id %s has limit %s %s, but the total debit amount is %s %s",
					uid,
					limitAmount.Amount().String(),
					limitAmount.CurrencyCode(),
					totalCredit.ItemAmount.String(),
					totalCredit.ItemCurrencyCode,
				)
				m.logger.Info(err.Error())
				return err
			}
			return err
		}
	}
	return nil
}

func (m *maxCreditPerTransfer) Name() string {
	return LimitMaxCreditPerTransfer
}
