package scheduled_transaction

import (
	"errors"

	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type scheduleValidator func(params *ScheduleParams, repo *Repository, logger log15.Logger) error

var (
	validatorAmountIsNegative = func(
		params *ScheduleParams,
		repo *Repository,
		logger log15.Logger,
	) error {
		if params.Amount.GreaterThanOrEqual(decimal.Zero) {
			return ErrorInvalidAmount
		}
		return nil
	}

	validatorAmountIsPositive = func(
		params *ScheduleParams,
		repo *Repository,
		logger log15.Logger,
	) error {
		if params.Amount.LessThanOrEqual(decimal.Zero) {
			return ErrorInvalidAmount
		}
		return nil
	}

	validatorTransactionIsNotExist = func(
		params *ScheduleParams,
		repo *Repository,
		logger log15.Logger,
	) error {
		logger = logger.New("validator", "validatorTransactionIsNotExist")

		_, err := repo.FindNextPendingByAccountIdAndReason(params.Account.ID, params.Reason)
		if err == gorm.ErrRecordNotFound {
			return nil
		}

		if err != nil {
			logger.Error("failed to find scheduled transaction", "error", err, "accountId", params.Account.ID)
			return err
		}
		return ErrorAlreadyScheduled
	}

	validatorChain = func(validators ...scheduleValidator) scheduleValidator {
		return func(params *ScheduleParams, repo *Repository, logger log15.Logger) error {
			for _, v := range validators {
				err := v(params, repo, logger)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
)

var validators = map[Reason]scheduleValidator{
	ReasonMaintenanceFee:     validatorChain(validatorAmountIsNegative, validatorTransactionIsNotExist),
	ReasonCreditLineFee:      validatorChain(validatorAmountIsNegative),
	ReasonLimitBalanceFee:    validatorChain(validatorAmountIsNegative, validatorTransactionIsNotExist),
	ReasonInterestGeneration: validatorChain(validatorAmountIsPositive),
}

var (
	ErrorAlreadyScheduled = errors.New("transfer already scheduled")
	ErrorInvalidAmount    = errors.New("invalid amount value")
)
