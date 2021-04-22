package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	txModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-accounts/internal/transfer/builder"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// CreditAccount is used in order to perform "Credit Account" transfer
type CreditAccount struct {
	db               *gorm.DB
	input            CreditAccountInput
	currencyProvider transfer.CurrencyProvider

	transactionsContainer
}

// NewCreditAccount is a CreditAccount constructor
func NewCreditAccount(db *gorm.DB, input CreditAccountInput, currencyProvider transfer.CurrencyProvider) *CreditAccount {
	return &CreditAccount{
		db:               db,
		input:            input,
		currencyProvider: currencyProvider,
	}
}

func (c *CreditAccount) Execute(request *model.Request) (types.Details, error) {
	details, err := c.Evaluate(request)

	if err != nil {
		return nil, err
	}

	err = saveTransactions(c.db, c.Transactions(), txModel.StatusExecuted)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save transaction")
	}

	input := c.input
	account := input.Account()
	err = updateAccount(c.db, account)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update account #%d (%s)", account.ID, account.Number)
	}
	if input.ApplyIWTFee() || input.DebitFromRevenueAccount() {
		revenueAcc := input.RevenueAccount()
		err = updateRevenueAccount(c.db, revenueAcc)
		if err != nil {
			return nil, err
		}
	}
	err = updateRequestStatus(c.db, request, "executed")
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (c *CreditAccount) Evaluate(request *model.Request) (types.Details, error) {
	acc := c.input.Account()
	rev := c.input.RevenueAccount()
	currency, err := c.currencyProvider.Get(*request.BaseCurrencyCode)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to evaluate transfer, an error occurred while trying to get currency %s data",
			*request.BaseCurrencyCode,
		)
	}
	amount := transfer.NewAmount(currency, *request.Amount)
	details := make(map[constants.Purpose]*types.Detail)

	chain := builder.New()

	if c.input.DebitFromRevenueAccount() {
		chain.
			Debit(amount).
			From(makeDebitable(currency, &rev.Balance, &rev.AvailableAmount)).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					RevenueAccountId:         &rev.ID,
					Description:              pointer.ToString("debit account"),
					Amount:                   pointer.ToDecimal(action.Amount().Neg()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(rev.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(rev.Balance),
					Type:                     pointer.ToString("revenue"),
					Purpose:                  pointer.ToString(constants.PurposeDebitRevenue.String()),
				}
				c.appendTransaction(transaction)

				details[constants.PurposeDebitRevenue] = &types.Detail{
					Purpose:          constants.PurposeDebitRevenue,
					Amount:           action.Amount().Neg(),
					CurrencyCode:     currency.Code(),
					Transaction:      transaction,
					RevenueAccountId: &rev.ID,
					RevenueAccount:   rev,
				}
				return err
			})
	}

	accountCreditable := makeCreditable(currency, &acc.AvailableAmount, &acc.Balance)
	chain.
		Credit(amount).
		To(accountCreditable).
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			currency := action.Currency()
			transaction := &txModel.Transaction{
				RequestId:                request.Id,
				AccountId:                &acc.ID,
				Description:              pointer.ToString("credit account"),
				Amount:                   pointer.ToDecimal(action.Amount()),
				IsVisible:                pointer.ToBool(true),
				AvailableBalanceSnapshot: pointer.ToDecimal(acc.AvailableAmount),
				CurrentBalanceSnapshot:   pointer.ToDecimal(acc.Balance),
				Type:                     pointer.ToString("account"),
				Purpose:                  pointer.ToString(constants.PurposeCreditAccount.String()),
			}
			c.appendTransaction(transaction)

			details[constants.PurposeCreditAccount] = &types.Detail{
				Purpose:      constants.PurposeCreditAccount,
				Amount:       action.Amount(),
				CurrencyCode: currency.Code(),
				Transaction:  transaction,
				AccountId:    pointer.ToUint64(acc.ID),
				Account:      acc,
			}
			return err
		})
	if c.input.ApplyIWTFee() {
		params := c.input.FeeParams()
		feeAmount := fee.NewTransferFeeAmount(*params, amount)
		chain.
			Debit(feeAmount).
			From(makeDebitable(currency, &acc.AvailableAmount, &acc.Balance)).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					AccountId:                &acc.ID,
					Description:              pointer.ToString("Transfer Fee: IWT Fee"),
					Amount:                   pointer.ToDecimal(action.Amount().Neg()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(acc.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(acc.Balance),
					Type:                     pointer.ToString("fee"),
					Purpose:                  pointer.ToString(constants.PurposeFeeIWT.String()),
				}
				c.appendTransaction(transaction)

				details[constants.PurposeFeeIWT] = &types.Detail{
					Purpose:      constants.PurposeFeeIWT,
					Amount:       action.Amount(),
					CurrencyCode: currency.Code(),
					Transaction:  transaction,
					AccountId:    pointer.ToUint64(acc.ID),
					Account:      acc,
				}
				return err
			}).
			Credit(feeAmount).
			To(makeCreditable(currency, &rev.AvailableAmount, &rev.Balance)).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					RevenueAccountId:         &rev.ID,
					Description:              pointer.ToString("Transfer Fee: IWT Fee"),
					Amount:                   pointer.ToDecimal(action.Amount()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(rev.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(rev.Balance),
					Type:                     pointer.ToString("revenue"),
					Purpose:                  pointer.ToString(constants.PurposeRevenueIwt.String()),
				}
				c.appendTransaction(transaction)

				details[constants.PurposeRevenueIwt] = &types.Detail{
					Purpose:          constants.PurposeRevenueIwt,
					Amount:           action.Amount(),
					CurrencyCode:     currency.Code(),
					Transaction:      transaction,
					RevenueAccountId: &rev.ID,
					RevenueAccount:   rev,
				}
				return err
			})
	}

	err = chain.Execute()
	return details, err
}
