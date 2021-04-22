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
)

type DeductRevenueAccount struct {
	db               *gorm.DB
	input            DRAInput
	currencyProvider transfer.CurrencyProvider

	transactionsContainer
}

func NewDeductRevenueAccount(
	db *gorm.DB,
	input DRAInput,
	currencyProvider transfer.CurrencyProvider,
) *DeductRevenueAccount {
	return &DeductRevenueAccount{db: db, input: input, currencyProvider: currencyProvider}
}

func (d *DeductRevenueAccount) Execute(request *model.Request) (types.Details, error) {
	if request.Status == nil {
		return nil, errors.Wrap(ErrMissingRequestData, "request.Status is required")
	}

	if *request.Status != "new" {
		return nil, errors.Wrapf(
			ErrUnexpectedStatus,
			`only requests with status "new" could be executed: got "%s" status`,
			*request.Status,
		)
	}

	details, err := d.Evaluate(request)
	if err != nil {
		return nil, err
	}

	err = saveTransactions(d.db, d.Transactions(), txModel.StatusExecuted)
	if err != nil {
		return nil, err
	}

	revenueAccount, err := d.input.RevenueAccount()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve revenue account")
	}

	err = updateRevenueAccount(d.db, revenueAccount)
	if err != nil {
		return nil, err
	}

	err = updateRequestStatus(d.db, request, "executed")

	return details, err
}

func (d *DeductRevenueAccount) Evaluate(request *model.Request) (types.Details, error) {
	input := d.input
	revenueAccount, err := input.RevenueAccount()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve revenue account")
	}

	currency, err := d.currencyProvider.Get(revenueAccount.CurrencyCode)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"failed to retrieve currency(%s) for revenue account #%d",
			revenueAccount.CurrencyCode,
			revenueAccount.ID,
		)
	}

	details, err := d.evaluate(
		request,
		revenueAccount,
		makeDebitable(currency, &revenueAccount.Balance, &revenueAccount.AvailableAmount),
	)

	return details, err
}

func (d *DeductRevenueAccount) evaluate(
	request *model.Request,
	revenueAccount *accountModel.RevenueAccountModel,
	source transfer.Debitable,
) (types.Details, error) {
	currency := source.Currency()
	if request.BaseCurrencyCode == nil {
		return nil, errors.Wrap(ErrMissingRequestData, "request.BaseCurrencyCode is required")
	}
	if *request.BaseCurrencyCode != currency.Code() {
		return nil, errors.Wrapf(
			transfer.ErrCurrenciesMismatch,
			"revenue account currency %s does not match request base currency %s",
			currency.Code(),
			*request.BaseCurrencyCode,
		)
	}
	details := make(map[constants.Purpose]*types.Detail)
	chain := builder.
		Debit(request.Amount).
		From(source).WithCallback(func(action transfer.Action) error {
		err := action.Perform()
		currency := action.Currency()
		transaction := &txModel.Transaction{
			RequestId:                request.Id,
			RevenueAccountId:         pointer.ToUint64(revenueAccount.ID),
			Description:              request.Description,
			Amount:                   pointer.ToDecimal(action.Amount().Neg()),
			IsVisible:                pointer.ToBool(true),
			AvailableBalanceSnapshot: pointer.ToDecimal(revenueAccount.AvailableAmount),
			CurrentBalanceSnapshot:   pointer.ToDecimal(revenueAccount.Balance),
			Type:                     pointer.ToString("revenue"),
			Purpose:                  pointer.ToString(constants.PurposeDebitRevenue.String()),
		}
		d.appendTransaction(transaction)

		details[constants.PurposeDebitRevenue] = &types.Detail{
			Purpose:          constants.PurposeDebitRevenue,
			Amount:           action.Amount().Neg(),
			CurrencyCode:     currency.Code(),
			Transaction:      transaction,
			RevenueAccountId: pointer.ToUint64(revenueAccount.ID),
			RevenueAccount:   revenueAccount,
		}
		return err
	})
	err := chain.Execute()

	return details, err
}
