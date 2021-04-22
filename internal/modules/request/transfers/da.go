package transfers

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	txModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-accounts/internal/transfer/builder"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type DebitAccount struct {
	db               *gorm.DB
	input            DaInput
	currencyProvider transfer.CurrencyProvider

	transactionsContainer
}

func NewDebitAccount(db *gorm.DB, input DaInput, currencyProvider transfer.CurrencyProvider) *DebitAccount {
	return &DebitAccount{db: db, input: input, currencyProvider: currencyProvider}
}

func (d *DebitAccount) Execute(request *model.Request) (types.Details, error) {
	var revenueCreditable transfer.Creditable

	sourceAccount, creditToRevenue, revenueAccount, currency, err := d.extract(request)
	if err != nil {
		return nil, err
	}

	permissions, err := d.permissions(*request.Amount)
	if err != nil {
		return nil, err
	}

	err = permissions.Check()
	if err != nil {
		return nil, err
	}

	if creditToRevenue {
		revenueCreditable = makeCreditable(currency, &revenueAccount.Balance, &revenueAccount.AvailableAmount)
	}
	details, err := d.evaluate(
		request,
		makeDebitable(currency, &sourceAccount.Balance, &sourceAccount.AvailableAmount),
		revenueCreditable,
		sourceAccount,
		revenueAccount,
		creditToRevenue,
	)
	if err != nil {
		return nil, err
	}

	err = saveTransactions(d.db, d.transactions, "executed")
	if err != nil {
		return nil, err
	}

	err = updateAccount(d.db, sourceAccount)
	if err != nil {
		return nil, err
	}

	if creditToRevenue {
		err = updateRevenueAccount(d.db, revenueAccount)
		if err != nil {
			return nil, err
		}
	}

	err = updateRequestStatus(d.db, request, "executed")
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (d *DebitAccount) Evaluate(request *model.Request) (types.Details, error) {
	var revenueCreditable transfer.Creditable

	sourceAccount, creditToRevenue, revenueAccount, currency, err := d.extract(request)
	if err != nil {
		return nil, err
	}
	if creditToRevenue {
		revenueCreditable = makeCreditable(currency, &revenueAccount.Balance, &revenueAccount.AvailableAmount)
	}
	return d.evaluate(
		request,
		makeDebitable(currency, &sourceAccount.Balance, &sourceAccount.AvailableAmount),
		revenueCreditable,
		sourceAccount,
		revenueAccount,
		creditToRevenue,
	)
}

func (d *DebitAccount) extract(request *model.Request) (
	sourceAccount *accountModel.Account,
	creditToRevenue bool,
	revenueAccount *accountModel.RevenueAccountModel,
	currency transfer.Currency,
	err error,
) {
	input := d.input
	sourceAccount, err = input.SourceAccount()
	if err != nil {
		err = errors.Wrap(err, "failed to retrieve SourceAccount from input")
		return
	}

	creditToRevenue, err = input.CreditToRevenueAccount()
	if err != nil {
		err = errors.Wrap(err, "failed to retrieve CreditToRevenueAccount parameter from input")
		return
	}
	currency, _, err = currencies(d.currencyProvider, request)
	if err != nil {
		return
	}

	if creditToRevenue {
		revenueAccount, err = input.RevenueAccount()
		if err != nil {
			err = errors.Wrap(err, "failed to retrieve RevenueAccount from input")
			return
		}
	}

	return
}

func (d *DebitAccount) evaluate(
	request *model.Request,
	source transfer.Debitable,
	revenue transfer.Creditable,
	sourceAccount *accountModel.Account,
	revenueAccount *accountModel.RevenueAccountModel,
	creditToRevenue bool,
) (types.Details, error) {
	if request.Amount == nil {
		return nil, errors.Wrap(ErrMissingRequestData, "request.amount is required")
	}

	details := make(map[constants.Purpose]*types.Detail)
	chain := builder.
		Debit(request.Amount).
		From(source).
		As("debitAmount").
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			currency := action.Currency()
			transaction := &txModel.Transaction{
				RequestId:                request.Id,
				AccountId:                &sourceAccount.ID,
				Description:              request.Description,
				Amount:                   pointer.ToDecimal(action.Amount().Neg()),
				IsVisible:                pointer.ToBool(true),
				AvailableBalanceSnapshot: pointer.ToDecimal(sourceAccount.AvailableAmount),
				CurrentBalanceSnapshot:   pointer.ToDecimal(sourceAccount.Balance),
				Type:                     pointer.ToString("account"),
				Purpose:                  pointer.ToString(constants.PurposeDebitAccount.String()),
			}
			d.appendTransaction(transaction)

			details[constants.PurposeDebitAccount] = &types.Detail{
				Purpose:      constants.PurposeDebitAccount,
				Amount:       action.Amount().Neg(),
				CurrencyCode: currency.Code(),
				Transaction:  transaction,
				AccountId:    &sourceAccount.ID,
				Account:      sourceAccount,
			}
			return err
		})

	if creditToRevenue {
		chain.
			CreditFromAlias("debitAmount").
			To(revenue).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					RevenueAccountId:         &revenueAccount.ID,
					Description:              request.Description,
					Amount:                   pointer.ToDecimal(action.Amount()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(revenueAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(revenueAccount.Balance),
					Type:                     pointer.ToString("revenue"),
					Purpose:                  pointer.ToString(constants.PurposeCreditRevenue.String()),
				}
				d.appendTransaction(transaction)

				details[constants.PurposeCreditRevenue] = &types.Detail{
					Purpose:          constants.PurposeCreditRevenue,
					Amount:           action.Amount(),
					CurrencyCode:     currency.Code(),
					Transaction:      transaction,
					RevenueAccountId: &revenueAccount.ID,
					RevenueAccount:   revenueAccount,
				}
				return err
			})
	}

	err := chain.Execute()
	return details, err
}

func (d *DebitAccount) permissions(requestedAmount decimal.Decimal) (PermissionCheckers, error) {
	input := d.input
	sourceAccount, err := input.SourceAccount()
	if err != nil {
		return nil, err
	}
	permissions := PermissionCheckers{
		NewWithdrawalPermission(sourceAccount),
	}
	allowNeg, err := input.AllowNegativeBalance()
	if err != nil {
		return nil, err
	}
	if !allowNeg {
		permissions = append(
			permissions,
			NewSufficientBalancePermission(
				SimpleAmountable(requestedAmount),
				SimpleAmountable(sourceAccount.AvailableAmount),
			))
	}
	return permissions, nil
}
