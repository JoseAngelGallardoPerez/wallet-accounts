package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

const (
	// LimitMaxDebitPerTransfer is responsible for limiting the amount that the user
	// can debit within single transfer
	LimitMaxDebitPerTransfer                = "max_debit_per_transfer"
	LimitMaxDebitPerTransferEnabled         = true
	LimitMaxDebitPerTransferDefaultAmount   = .0
	LimitMaxDebitPerTransferDefaultCurrency = "EUR"
)

type maxDebitPerTransfer struct {
	details      types.Details
	limitService *limit.Service
	reducer      balance.Reducer
	logger       log15.Logger
}

func NewMaxDebitPerTransfer(
	details types.Details,
	limitService *limit.Service,
	reducer balance.Reducer,
	logger log15.Logger,
) PermissionChecker {
	return &maxDebitPerTransfer{
		details:      details,
		limitService: limitService,
		reducer:      reducer,
		logger:       logger,
	}
}

func (m *maxDebitPerTransfer) Check() error {
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
			Name:     LimitMaxDebitPerTransfer,
			Entity:   "user",
			EntityId: uid,
		})
		// if no limit found then no need to check
		if errors.Cause(err) == limit.ErrNotFound {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "failed to check max debit per transfer: limit service returned error")
		}
		if lim.Available().NoLimit() {
			continue
		}
		limitAmount := lim.Available().CurrencyAmount()
		totalDebit, err := m.reducer.Reduce(aggregation, limitAmount.CurrencyCode())
		if err != nil {
			return errors.Wrap(err, "failed to check max debit per transfer: balance reducer returned error")
		}

		if err = lim.WithinLimit(&totalDebit); err != nil {
			if errors.Cause(err) == limit.ErrLimitExceeded {
				// TODO: get rid of logger here, see https://velmie.atlassian.net/browse/VW-296
				err = errors.Wrapf(
					err,
					"max debit per transfer limit is exceeded: user with id %s has limit %s %s, but the total debit amount is %s %s",
					uid,
					limitAmount.Amount().String(),
					limitAmount.CurrencyCode(),
					totalDebit.ItemAmount.String(),
					totalDebit.ItemCurrencyCode,
				)
				m.logger.Info(err.Error())
				return err
			}
			return err
		}
	}
	return nil
}

func (m *maxDebitPerTransfer) Name() string {
	return LimitMaxDebitPerTransfer
}
