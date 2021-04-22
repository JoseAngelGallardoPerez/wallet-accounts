package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/exchange"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	txModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/Confialink/wallet-accounts/internal/transfer/builder"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type OutgoingWireTransfer struct {
	input             OWTInput
	currencyProvider  transfer.CurrencyProvider
	db                *gorm.DB
	permissionFactory PermissionFactory

	transactionsContainer
}

func NewOutgoingWireTransfer(
	input OWTInput,
	currencyProvider transfer.CurrencyProvider,
	db *gorm.DB,
	pf PermissionFactory,
) *OutgoingWireTransfer {
	return &OutgoingWireTransfer{
		input:             input,
		currencyProvider:  currencyProvider,
		db:                db,
		permissionFactory: pf.WrapContext(db),
	}
}

func owTransfer(
	db *gorm.DB,
	request *model.Request,
	provider transfer.CurrencyProvider,
	permissionFactory PermissionFactory,
) *OutgoingWireTransfer {
	input := NewDbOWTInput(db, request, nil)
	return NewOutgoingWireTransfer(input, provider, db, permissionFactory)
}

func (o *OutgoingWireTransfer) Evaluate(request *model.Request) (types.Details, error) {
	input := o.input
	sourceAccount, err := input.SourceAccount()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve source account, request id = %d", request.Id)
	}

	revenueAccount, err := input.RevenueAccount()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve revenue account, request id = %d", request.Id)
	}

	baseCurrency, referenceCurrency, err := currencies(o.currencyProvider, request)
	if err != nil {
		return nil, err
	}

	return o.evaluate(
		request,
		sourceAccount,
		revenueAccount,
		makeDebitable(baseCurrency, &sourceAccount.AvailableAmount, &sourceAccount.Balance),
		makeCreditable(baseCurrency, &revenueAccount.AvailableAmount, &revenueAccount.Balance),
		baseCurrency,
		referenceCurrency,
	)
}

func (o *OutgoingWireTransfer) DryRun(request *model.Request) (types.Details, error) {
	sourceAccount, err := o.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	revenueAccount, err := o.input.RevenueAccount()
	if err != nil {
		return nil, err
	}
	baseCurrency, referenceCurrency, err := currencies(o.currencyProvider, request)
	if err != nil {
		return nil, err
	}

	sourceNoop := transfer.NewNoOpWallet(baseCurrency)
	revenueNoop := transfer.NewNoOpWallet(baseCurrency)
	return o.evaluate(
		request,
		sourceAccount,
		revenueAccount,
		sourceNoop,
		revenueNoop,
		baseCurrency,
		referenceCurrency,
	)
}

func (o *OutgoingWireTransfer) Pending(request *model.Request) (types.Details, error) {
	if *request.Status != "new" {
		return nil, errors.Wrapf(ErrUnexpectedStatus, "expected status new, but got %s", *request.Status)
	}

	details, err := o.DryRun(request)
	if err != nil {
		return nil, err
	}
	permissions, err := o.permissionFactory.CreatePermission(request, details)
	if err != nil {
		return nil, err
	}
	err = permissions.Check()
	if err != nil {
		return nil, err
	}

	input := o.input
	sourceAccount, err := input.SourceAccount()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve source account, request id = %d", request.Id)
	}

	revenueAccount, err := input.RevenueAccount()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve revenue account, request id = %d", request.Id)
	}

	baseCurrency, referenceCurrency, err := currencies(o.currencyProvider, request)
	if err != nil {
		return nil, err
	}
	baseNoOp := transfer.NewNoOpWallet(baseCurrency)

	details, err = o.evaluate(
		request,
		sourceAccount,
		revenueAccount,
		makeDebitable(baseCurrency, &sourceAccount.AvailableAmount, nil),
		baseNoOp, // instead of revenue account b/c pending request should not impact revenue account balance
		baseCurrency,
		referenceCurrency,
	)

	if err != nil {
		return nil, err
	}

	err = saveTransactions(o.db, o.transactions, "pending")
	if err != nil {
		return nil, err
	}

	err = updateAccount(o.db, sourceAccount)
	if err != nil {
		return nil, err
	}

	err = updateRequestStatusAndAmount(o.db, request, "pending")
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (o *OutgoingWireTransfer) Execute(request *model.Request) (types.Details, error) {
	switch *request.Status {
	case "new":
		return o.executeNewRequest(request)
	case "pending":
		return o.executePendingRequest(request)
	}
	return nil, errors.Wrapf(
		ErrUnexpectedStatus,
		`request could be executed from status "new" or "pending": got "%s" status`,
		*request.Status,
	)
}

func (o *OutgoingWireTransfer) Cancel(request *model.Request, reason string) error {
	if *request.Status != "pending" {
		return errors.Wrapf(
			ErrUnexpectedStatus,
			`only requests with status "pending" could be cancelled: got "%s" status`,
			*request.Status,
		)
	}

	transactions, err := loadTransactions(o.db, *request.Id)
	if err != nil {
		return err
	}
	sourceAccount, err := o.input.SourceAccount()
	if err != nil {
		return err
	}
	o.transactions = transactions
	for _, t := range transactions {
		t.Status = pointer.ToString("cancelled")
		if t.AccountId != nil && *t.AccountId == sourceAccount.ID {
			sourceAccount.AvailableAmount = sourceAccount.AvailableAmount.Add(t.Amount.Neg())
		}
	}

	err = updateTransactionsStatusByRequestId(o.db, *request.Id, "cancelled")
	if err != nil {
		return err
	}
	// update source and destination account balances
	err = updateAccount(o.db, sourceAccount)
	if err != nil {
		return err
	}
	request.CancellationReason = &reason
	err = updateRequestStatusAndCancellationReason(o.db, request, "cancelled", reason)
	if err != nil {
		return err
	}
	return nil
}

func (o *OutgoingWireTransfer) Modify(request *model.Request) (types.Details, error) {
	if *request.Status != "pending" {
		return nil, errors.Wrapf(
			ErrUnexpectedStatus,
			`only requests with status "pending" could be modified: got "%s" status`,
			*request.Status,
		)
	}
	// load "pending" transactions
	transactions, err := loadTransactions(o.db, *request.Id)
	if err != nil {
		return nil, err
	}

	input := o.input
	sourceAccount, err := input.SourceAccount()
	if err != nil {
		return nil, err
	}

	revenueAccount, err := input.RevenueAccount()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve revenue account, request id = %d", request.Id)
	}

	// restore source account available amount
	for _, t := range transactions {
		if t.AccountId != nil && *t.AccountId == sourceAccount.ID {
			sourceAccount.AvailableAmount = sourceAccount.AvailableAmount.Add(t.Amount.Neg())
		}
	}

	baseCurrency, referenceCurrency, err := currencies(o.currencyProvider, request)
	baseNoOp := transfer.NewNoOpWallet(baseCurrency)
	if err != nil {
		return nil, err
	}
	// evaluate request with updated request rate (re-calculate)
	details, err := o.evaluate(
		request,
		sourceAccount,
		revenueAccount,
		makeDebitable(baseCurrency, &sourceAccount.AvailableAmount, nil),
		baseNoOp, // instead of revenue account b/c pending request should not impact revenue account balance
		baseCurrency,
		referenceCurrency,
	)
	if err != nil {
		return nil, err
	}
	if len(details) != len(transactions) {
		return nil, errors.Wrap(
			ErrModificationNotAllowed,
			"The number of transactions in the request has changed. It is assumed that changes will only affect existing transactions.",
		)
	}
	// update transactions
	err = syncAndUpdateTransactions(o.db, details, transactions, txModel.StatusPending)
	if err != nil {
		return nil, err
	}
	// in OWT request amount is recalculated b/c input amount is specified in reference currency
	err = updateRequestAmountAndRate(o.db, request)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update request amount(%s) #%d", request.Amount, *request.Id)
	}

	err = updateAccount(o.db, sourceAccount)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (o *OutgoingWireTransfer) executeNewRequest(request *model.Request) (types.Details, error) {
	details, err := o.DryRun(request)
	if err != nil {
		return nil, err
	}
	permissions, err := o.permissionFactory.CreatePermission(request, details)
	if err != nil {
		return nil, err
	}
	err = permissions.Check()
	if err != nil {
		return nil, err
	}

	details, err = o.Evaluate(request)
	if err != nil {
		return nil, err
	}
	err = saveTransactions(o.db, o.Transactions(), txModel.StatusExecuted)
	if err != nil {
		return nil, err
	}
	sourceAccount, err := o.input.SourceAccount()
	if err != nil {
		return nil, err
	}

	revenueAccount, err := o.input.RevenueAccount()
	if err != nil {
		return nil, err
	}

	err = updateAccount(o.db, sourceAccount)
	if err != nil {
		return nil, err
	}
	err = updateRevenueAccount(o.db, revenueAccount)
	if err != nil {
		return nil, err
	}
	err = updateRequestStatus(o.db, request, "executed")

	return details, err
}

func (o *OutgoingWireTransfer) executePendingRequest(request *model.Request) (types.Details, error) {
	transactions, err := loadTransactions(o.db, *request.Id)
	if err != nil {
		return nil, err
	}
	sourceAccount, err := o.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	revenueAccount, err := o.input.RevenueAccount()
	if err != nil {
		return nil, err
	}
	baseCurrency, referenceCurrency, err := currencies(o.currencyProvider, request)
	if err != nil {
		return nil, err
	}
	details, err := o.evaluate(
		request,
		sourceAccount,
		revenueAccount,
		// only change balance because available amount has been already changed during "Pending" operation
		makeDebitable(baseCurrency, &sourceAccount.Balance, nil),
		makeCreditable(baseCurrency, &revenueAccount.Balance, &revenueAccount.AvailableAmount),
		baseCurrency,
		referenceCurrency,
	)
	if err != nil {
		return nil, err
	}

	o.transactions = transactions

	err = syncAndUpdateTransactions(o.db, details, transactions, txModel.StatusExecuted)
	if err != nil {
		return nil, err
	}

	err = updateAccount(o.db, sourceAccount)
	if err != nil {
		return nil, err
	}

	err = updateRevenueAccount(o.db, revenueAccount)
	if err != nil {
		return nil, err
	}
	err = updateRequestStatus(o.db, request, "executed")
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (o *OutgoingWireTransfer) evaluate(
	request *model.Request,
	sourceAccount *accountModel.Account,
	revenueAccount *accountModel.RevenueAccountModel,
	source transfer.Debitable,
	revenue transfer.Creditable,
	baseCurrency transfer.Currency,
	referenceCurrency transfer.Currency,
) (types.Details, error) {
	o.transactions = nil
	if request.Rate == nil {
		return nil, errors.Wrap(ErrMissingRequestData, "request.Rate is required")
	}
	if request.InputAmount == nil {
		return nil, errors.Wrap(ErrMissingRequestData, "request.InputAmount is required")
	}
	rateSource := exchange.NewDirectRateSource()
	_ = rateSource.Set(exchange.NewRate(request.RateBaseCurrencyCode(), request.RateReferenceCurrencyCode(), *request.Rate))

	input := o.input
	details := make(map[constants.Purpose]*types.Detail)
	inputAmount := transfer.NewAmount(referenceCurrency, request.GetInputAmount())
	//todo: probably it makes sense to move such conversions out (make common layer)
	convert := builder.
		Exchange(inputAmount).
		Using(rateSource).
		ToCurrency(baseCurrency).
		As("convertedAmount").
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			convertedAmount := action.Amount()
			request.Amount = &convertedAmount
			return err
		})
	err := convert.Execute()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert requested amount")
	}
	convertedAmount := convert.AmountAlias("convertedAmount")

	exchangeMarginPercent, err := input.ExchangeMarginPercent()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve exchange margin, request id = %d", request.Id)
	}

	chain := builder.New()
	marginFeeApplied := false
	exchangeMarginMultiplier := exchangeMarginPercent.Div(decimal.NewFromInt(100))
	if exchangeMarginMultiplier.GreaterThan(decimal.NewFromInt(0)) {
		marginFeeApplied = true
		exchangeMarginAmount := transfer.NewAmountMultiplier(convertedAmount, exchangeMarginMultiplier)
		chain.
			Debit(exchangeMarginAmount).
			From(source).
			IncludeToGroup("showAmount").
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					AccountId:                &sourceAccount.ID,
					Description:              pointer.ToString("Conversion Margin"),
					Amount:                   pointer.ToDecimal(action.Amount().Neg()),
					IsVisible:                pointer.ToBool(false),
					AvailableBalanceSnapshot: pointer.ToDecimal(sourceAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(sourceAccount.Balance),
					Type:                     pointer.ToString("fee"),
					Purpose:                  pointer.ToString(constants.PurposeFeeExchangeMargin.String()),
				}
				o.appendTransaction(transaction)

				details[constants.PurposeFeeExchangeMargin] = &types.Detail{
					Purpose:      constants.PurposeFeeExchangeMargin,
					Amount:       action.Amount().Neg(),
					CurrencyCode: currency.Code(),
					Transaction:  transaction,
					AccountId:    &sourceAccount.ID,
					Account:      sourceAccount,
				}
				return err
			}).
			As("exchangeMargin")
	}

	chain.
		Debit(convertedAmount).
		From(source).
		IncludeToGroup("showAmount").
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			currency := action.Currency()
			if err != nil {
				return err
			}
			outgoingDescription, err := o.outgoingDescription(request, action.Amount())
			if err != nil {
				return err
			}
			transaction := &txModel.Transaction{
				RequestId:                request.Id,
				AccountId:                &sourceAccount.ID,
				Description:              &outgoingDescription,
				Amount:                   pointer.ToDecimal(action.Amount().Neg()),
				IsVisible:                pointer.ToBool(true),
				AvailableBalanceSnapshot: pointer.ToDecimal(sourceAccount.AvailableAmount),
				CurrentBalanceSnapshot:   pointer.ToDecimal(sourceAccount.Balance),
				Type:                     pointer.ToString("account"),
				Purpose:                  pointer.ToString(constants.PurposeOWTOutgoing.String()),
			}
			// this group include exchange margin and outgoing value
			// if exchange margin is not apply then showAmount equals outgoing amount
			// it is needed because exchange margin is not visible (it shown as included in outgoing transaction)
			showAmount := chain.GetGroup("showAmount").Sum()
			if !showAmount.Equal(*transaction.Amount) {
				transaction.ShowAmount = &showAmount
			}
			o.appendTransaction(transaction)

			details[constants.PurposeOWTOutgoing] = &types.Detail{
				Purpose:      constants.PurposeOWTOutgoing,
				Amount:       action.Amount().Neg(),
				CurrencyCode: currency.Code(),
				Transaction:  transaction,
				AccountId:    &sourceAccount.ID,
				Account:      sourceAccount,
			}
			return nil
		})

	feeParams, err := input.TransferFeeParams()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve transfer fee params, request id = %d", request.Id)
	}

	if feeParams != nil {
		feeAmount := fee.NewTransferFeeAmount(*feeParams, convertedAmount)
		description := "Transfer Fee: OWT Fee"
		chain.
			Debit(feeAmount).
			From(source).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				if err != nil {
					return err
				}
				if action.Amount().IsZero() {
					return nil
				}
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					AccountId:                &sourceAccount.ID,
					Description:              &description,
					Amount:                   pointer.ToDecimal(action.Amount().Neg()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(sourceAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(sourceAccount.Balance),
					Type:                     pointer.ToString("fee"),
					Purpose:                  pointer.ToString(constants.PurposeFeeTransfer.String()),
				}
				o.appendTransaction(transaction)

				details[constants.PurposeFeeTransfer] = &types.Detail{
					Purpose:      constants.PurposeFeeTransfer,
					Amount:       action.Amount().Neg(),
					CurrencyCode: currency.Code(),
					Transaction:  transaction,
					AccountId:    pointer.ToUint64(sourceAccount.ID),
					Account:      sourceAccount,
				}
				return nil
			}).
			Credit(feeAmount).
			To(revenue).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				if err != nil {
					return err
				}
				if action.Amount().IsZero() {
					return nil
				}
				purpose := constants.Purpose("revenue_owt_transfer")
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					RevenueAccountId:         &revenueAccount.ID,
					Description:              &description,
					Amount:                   pointer.ToDecimal(action.Amount()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(revenueAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(revenueAccount.Balance),
					Type:                     pointer.ToString("revenue"),
					Purpose:                  pointer.ToString(purpose.String()),
				}
				o.appendTransaction(transaction)

				details[purpose] = &types.Detail{
					Purpose:          purpose,
					Amount:           action.Amount(),
					CurrencyCode:     currency.Code(),
					Transaction:      transaction,
					RevenueAccountId: &revenueAccount.ID,
					RevenueAccount:   revenueAccount,
				}
				return nil
			})
	}

	if marginFeeApplied {
		chain.
			CreditFromAlias("exchangeMargin").
			To(revenue).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					RevenueAccountId:         &revenueAccount.ID,
					Description:              pointer.ToString("Conversion margin"),
					Amount:                   pointer.ToDecimal(action.Amount()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(revenueAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(revenueAccount.Balance),
					Type:                     pointer.ToString("revenue"),
					Purpose:                  pointer.ToString(constants.PurposeRevenueExchangeMargin.String()),
				}
				o.appendTransaction(transaction)

				details[constants.PurposeRevenueExchangeMargin] = &types.Detail{
					Purpose:          constants.PurposeRevenueExchangeMargin,
					Amount:           action.Amount(),
					CurrencyCode:     currency.Code(),
					Transaction:      transaction,
					RevenueAccountId: &revenueAccount.ID,
					RevenueAccount:   revenueAccount,
				}
				return err
			})
	}

	return details, chain.Execute()
}

func (o *OutgoingWireTransfer) outgoingDescription(request *model.Request, outgoingAmount decimal.Decimal) (string, error) {
	d := "Outgoing Wire Transfer"

	if request.Description != nil && *request.Description != "" {
		d = *request.Description
	}

	input := o.input
	accName, err := input.BeneficiaryCustomerAccountName()
	if err != nil {
		return "",
			errors.Wrapf(
				err,
				"unable to compose outgoing transaction description: failed to retrieve beneficiary customer account name, request id = %d",
				request.Id,
			)
	}
	refMessage, err := input.RefMessage()
	if err != nil {
		return "",
			errors.Wrapf(
				err,
				"unable to compose outgoing transaction description: failed to retrieve ref message, request id = %d",
				request.Id,
			)
	}

	//TODO: Discuss with Alex P. and Elena E. how the description should look like. Before the hotfix - description contains a big number after point

	return fmt.Sprintf("%s %s %s %s %s",
		d,
		request.RateReferenceCurrencyCode(),
		outgoingAmount.Round(2).String(),
		accName,
		refMessage,
	), nil
}
