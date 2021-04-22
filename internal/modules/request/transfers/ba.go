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
	"strings"
)

type BetweenAccounts struct {
	subject           string
	currencyProvider  transfer.CurrencyProvider
	input             BetweenAccountsInput
	db                *gorm.DB
	permissionFactory PermissionFactory
	transactionsContainer
}

func NewBetweenAccounts(
	subject string,
	currencyProvider transfer.CurrencyProvider,
	input BetweenAccountsInput,
	db *gorm.DB,
	pf PermissionFactory,
) *BetweenAccounts {
	return &BetweenAccounts{
		subject:           subject,
		currencyProvider:  currencyProvider,
		input:             input,
		db:                db,
		permissionFactory: pf.WrapContext(db),
	}
}

// baTransfer creates transfer between accounts service with input that loads all required data by itself
func baTransfer(db *gorm.DB, request *model.Request, provider transfer.CurrencyProvider, pf PermissionFactory) *BetweenAccounts {
	input := NewDbBetweenAccountsInput(db, request, nil)
	return NewBetweenAccounts(request.Subject.String(), provider, input, db, pf)
}

func (b *BetweenAccounts) Evaluate(request *model.Request) (types.Details, error) {
	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	destinationAccount, err := b.input.DestinationAccount()
	if err != nil {
		return nil, err
	}
	revenueAccount, err := b.input.RevenueAccount()
	if err != nil {
		return nil, err
	}
	baseCurrency, referenceCurrency, err := currencies(b.currencyProvider, request)
	if err != nil {
		return nil, err
	}
	return b.evaluate(
		request,
		makeDebitable(baseCurrency, &sourceAccount.Balance, &sourceAccount.AvailableAmount),
		makeCreditable(referenceCurrency, &destinationAccount.Balance, &destinationAccount.AvailableAmount),
		makeCreditable(baseCurrency, &revenueAccount.Balance, &revenueAccount.AvailableAmount),
	)
}

func (b *BetweenAccounts) Pending(request *model.Request) (types.Details, error) {
	if *request.Status != "new" {
		return nil, errors.Wrapf(ErrUnexpectedStatus, "expected status new, but got %s", *request.Status)
	}
	details, err := b.DryRun(request)
	if err != nil {
		return nil, err
	}
	permissions, err := b.permissionFactory.CreatePermission(request, details)
	if err != nil {
		return nil, err
	}
	err = permissions.Check()
	if err != nil {
		return nil, err
	}

	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	baseCurrency, referenceCurrency, err := currencies(b.currencyProvider, request)
	baseNoOp := transfer.NewNoOpWallet(baseCurrency)
	referenceNoOp := transfer.NewNoOpWallet(referenceCurrency)
	if err != nil {
		return nil, err
	}
	details, err = b.evaluate(
		request,
		makeDebitable(baseCurrency, &sourceAccount.AvailableAmount, nil), // change only available amount
		referenceNoOp, // no operation
		baseNoOp,      // no operation
	)
	if err != nil {
		return nil, err
	}

	err = saveTransactions(b.db, b.Transactions(), txModel.StatusPending)
	if err != nil {
		return nil, err
	}
	err = b.db.
		Table(sourceAccount.TableName()).
		Where("id = ?", sourceAccount.ID).
		Updates(&accountModel.AccountPrivate{
			AvailableAmount: sourceAccount.AvailableAmount,
		}).Error
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update account #%d (%s)", sourceAccount.ID, sourceAccount.Number)
	}
	err = updateRequestStatus(b.db, request, "pending")
	if err != nil {
		return nil, err
	}
	return details, err
}

func (b *BetweenAccounts) Execute(request *model.Request) (types.Details, error) {
	switch *request.Status {
	case "new":
		return b.executeNewRequest(request)
	case "pending":
		return b.executePendingRequest(request)
	}
	return nil, errors.Wrapf(
		ErrUnexpectedStatus,
		`request could be executed from status "new" or "pending": got "%s" status`,
		*request.Status,
	)
}

func (b *BetweenAccounts) Modify(request *model.Request) (types.Details, error) {
	if *request.Status != "pending" {
		return nil, errors.Wrapf(
			ErrUnexpectedStatus,
			`only requests with status "pending" could be modified: got "%s" status`,
			*request.Status,
		)
	}
	// load "pending" transactions
	transactions, err := loadTransactions(b.db, *request.Id)
	if err != nil {
		return nil, err
	}
	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return nil, err
	}

	// restore source account available amount
	for _, t := range transactions {
		if t.AccountId != nil && *t.AccountId == sourceAccount.ID {
			sourceAccount.AvailableAmount = sourceAccount.AvailableAmount.Add(t.Amount.Neg())
		}
	}

	baseCurrency, referenceCurrency, err := currencies(b.currencyProvider, request)
	baseNoOp := transfer.NewNoOpWallet(baseCurrency)
	referenceNoOp := transfer.NewNoOpWallet(referenceCurrency)
	if err != nil {
		return nil, err
	}
	// evaluate request with updated request rate (re-calculate)
	details, err := b.evaluate(
		request,
		makeDebitable(baseCurrency, &sourceAccount.AvailableAmount, nil), // change only available amount
		referenceNoOp, // no operation
		baseNoOp,      // no operation
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
	err = syncAndUpdateTransactions(b.db, details, transactions, txModel.StatusPending)
	if err != nil {
		return nil, err
	}

	err = updateRequestAmountAndRate(b.db, request)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update request amount(%s) #%d", request.Amount, *request.Id)
	}

	return details, err
}

func (b *BetweenAccounts) Cancel(request *model.Request, reason string) error {
	if *request.Status != "pending" {
		return errors.Wrapf(
			ErrUnexpectedStatus,
			`only requests with status "pending" could be cancelled: got "%s" status`,
			*request.Status,
		)
	}

	transactions, err := loadTransactions(b.db, *request.Id)
	if err != nil {
		return err
	}
	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return err
	}
	b.transactions = transactions
	for _, t := range transactions {
		t.Status = pointer.ToString("cancelled")
		if t.AccountId != nil && *t.AccountId == sourceAccount.ID {
			sourceAccount.AvailableAmount = sourceAccount.AvailableAmount.Add(t.Amount.Neg())
		}
	}

	err = updateTransactionsStatusByRequestId(b.db, *request.Id, txModel.StatusCancelled)
	if err != nil {
		return err
	}
	// update source and destination account balances
	err = updateAccount(b.db, sourceAccount)
	if err != nil {
		return err
	}
	request.CancellationReason = &reason
	err = updateRequestStatusAndCancellationReason(b.db, request, txModel.StatusCancelled, reason)
	if err != nil {
		return err
	}
	return nil
}

func (b *BetweenAccounts) executeNewRequest(request *model.Request) (types.Details, error) {
	details, err := b.DryRun(request)
	if err != nil {
		return nil, err
	}
	permissions, err := b.permissionFactory.CreatePermission(request, details)
	if err != nil {
		return nil, err
	}
	err = permissions.Check()
	if err != nil {
		return nil, err
	}
	details, err = b.Evaluate(request)
	if err != nil {
		return nil, err
	}
	err = saveTransactions(b.db, b.Transactions(), txModel.StatusExecuted)
	if err != nil {
		return nil, err
	}
	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	destinationAccount, err := b.input.DestinationAccount()
	if err != nil {
		return nil, err
	}
	revenueAccount, err := b.input.RevenueAccount()
	if err != nil {
		return nil, err
	}

	// update source and destination account balances
	err = updateAccount(b.db, sourceAccount, destinationAccount)
	if err != nil {
		return nil, err
	}
	err = updateRevenueAccount(b.db, revenueAccount)
	if err != nil {
		return nil, err
	}
	err = updateRequestStatus(b.db, request, txModel.StatusExecuted)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (b *BetweenAccounts) executePendingRequest(request *model.Request) (types.Details, error) {
	transactions, err := loadTransactions(b.db, *request.Id)
	if err != nil {
		return nil, err
	}
	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	destinationAccount, err := b.input.DestinationAccount()
	if err != nil {
		return nil, err
	}
	revenueAccount, err := b.input.RevenueAccount()
	if err != nil {
		return nil, err
	}
	baseCurrency, referenceCurrency, err := currencies(b.currencyProvider, request)
	if err != nil {
		return nil, err
	}
	details, err := b.evaluate(
		request,
		// only change balance because available amount has been already changed during "Pending" operation
		makeDebitable(baseCurrency, &sourceAccount.Balance, nil),
		makeCreditable(referenceCurrency, &destinationAccount.Balance, &destinationAccount.AvailableAmount),
		makeCreditable(baseCurrency, &revenueAccount.Balance, &revenueAccount.AvailableAmount),
	)
	if err != nil {
		return nil, err
	}

	b.transactions = transactions

	err = syncAndUpdateTransactions(b.db, details, transactions, txModel.StatusExecuted)
	if err != nil {
		return nil, err
	}

	// update source and destination account balances
	err = updateAccount(b.db, sourceAccount, destinationAccount)
	if err != nil {
		return nil, err
	}

	err = updateRevenueAccount(b.db, revenueAccount)
	if err != nil {
		return nil, err
	}
	err = updateRequestStatus(b.db, request, "executed")
	if err != nil {
		return nil, err
	}

	return details, nil
}

// DryRun evaluates transfer but does not change accounts balance
func (b *BetweenAccounts) DryRun(request *model.Request) (types.Details, error) {
	baseCurrency, referenceCurrency, err := currencies(b.currencyProvider, request)
	if err != nil {
		return nil, err
	}
	sourceNoop, destinationNoop := transfer.NewNoOpWallet(baseCurrency), transfer.NewNoOpWallet(referenceCurrency)
	revenueNoop := transfer.NewNoOpWallet(baseCurrency)
	return b.evaluate(request, sourceNoop, destinationNoop, revenueNoop)
}

func (b *BetweenAccounts) evaluate(
	request *model.Request,
	source transfer.Debitable,
	destination transfer.Creditable,
	revenue transfer.Creditable,
) (types.Details, error) {
	b.transactions = nil
	sourceAccount, err := b.input.SourceAccount()
	if err != nil {
		return nil, err
	}
	sourceCurrency := source.Currency()
	destinationAccount, err := b.input.DestinationAccount()
	if err != nil {
		return nil, err
	}
	destinationCurrency := destination.Currency()
	revenueAccount, err := b.input.RevenueAccount()
	if err != nil {
		return nil, err
	}
	if sourceAccount.Type == nil {
		return nil, errors.New("source account type is nil, account type is required")
	}
	if destinationAccount.Type == nil {
		return nil, errors.New("destination account type is nil, account type is required")
	}
	if sourceAccount.Type.CurrencyCode != *request.BaseCurrencyCode {
		return nil, errors.Wrapf(
			transfer.ErrCurrenciesMismatch,
			"source account currency code (%s) must be the same as request base currency code (%s)",
			sourceAccount.Type.CurrencyCode,
			*request.BaseCurrencyCode,
		)
	}
	if destinationAccount.Type.CurrencyCode != *request.ReferenceCurrencyCode {
		return nil, errors.Wrapf(
			transfer.ErrCurrenciesMismatch,
			"destination account currency code (%s) must be the same as request reference currency code (%s)",
			destinationAccount.Type.CurrencyCode,
			*request.BaseCurrencyCode,
		)
	}
	if revenueAccount.CurrencyCode != *request.BaseCurrencyCode {
		return nil, errors.Wrapf(
			transfer.ErrCurrenciesMismatch,
			"revenue account currency code (%s) must be the same as request base currency code (%s)",
			sourceAccount.Type.CurrencyCode,
			*request.BaseCurrencyCode,
		)
	}

	requestAmount := *request.Amount
	debitAmount := transfer.NewAmount(source.Currency(), requestAmount)
	exchangeMarginPercent, err := b.input.ExchangeMarginPercent()
	if err != nil {
		return nil, err
	}
	// this is debit amount that will be used in order to split debit amount and exchange margin
	remainder := transfer.NewAmountConsumable(transfer.NewAmount(source.Currency(), requestAmount))
	exchangeMarginMultiplier := exchangeMarginPercent.Div(decimal.NewFromInt(100))

	chain := builder.New()
	details := make(map[constants.Purpose]*types.Detail)
	debitPurpose := b.debitPurpose()
	if exchangeMarginMultiplier.GreaterThan(decimal.NewFromInt(0)) {
		// join debitable in order to call debit from both initial amount and source account
		sourceAndRemainder, err := transfer.JoinDebitable(source, remainder)
		if err != nil {
			return nil, err
		}
		exchangeMarginAmount := transfer.NewAmountMultiplier(debitAmount, exchangeMarginMultiplier)
		chain.
			Debit(exchangeMarginAmount).
			From(sourceAndRemainder).
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
				b.appendTransaction(transaction)

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

	var outgoingDescription, incomingDescription string
	if request.Description != nil && *request.Description != "" {
		outgoingDescription, incomingDescription = *request.Description, *request.Description
	} else {
		outgoingDescription = fmt.Sprintf(`Transfer to "%s"`, destinationAccount.Number)
		incomingDescription = fmt.Sprintf(`Transfer from "%s"`, sourceAccount.Number)
	}
	chain.
		Debit(remainder).
		From(source).
		IncludeToGroup("showAmount").
		WithCallback(func(action transfer.Action) error {
			err := action.Perform()
			currency := action.Currency()
			transaction := &txModel.Transaction{
				RequestId:                request.Id,
				AccountId:                &sourceAccount.ID,
				Description:              &outgoingDescription,
				Amount:                   pointer.ToDecimal(action.Amount().Neg()),
				IsVisible:                pointer.ToBool(true),
				AvailableBalanceSnapshot: pointer.ToDecimal(sourceAccount.AvailableAmount),
				CurrentBalanceSnapshot:   pointer.ToDecimal(sourceAccount.Balance),
				Type:                     pointer.ToString("account"),
				Purpose:                  pointer.ToString(debitPurpose.String()),
			}
			// this group include exchange margin and outgoing value
			// if exchange margin is not apply then showAmount equals outgoing amount
			// it is needed because exchange margin is not visible (it shown as included in outgoing transaction)
			showAmount := chain.GetGroup("showAmount").Sum()
			if !showAmount.Equal(*transaction.Amount) {
				transaction.ShowAmount = &showAmount
			}

			b.appendTransaction(transaction)

			details[debitPurpose] = &types.Detail{
				Purpose:      debitPurpose,
				Amount:       action.Amount().Neg(),
				CurrencyCode: currency.Code(),
				Transaction:  transaction,
				AccountId:    &sourceAccount.ID,
				Account:      sourceAccount,
			}
			return err
		})

	transferFeeParams, err := b.input.TransferFeeParams()
	if err != nil {
		return nil, err
	}
	transferFeeDescription := b.transferFeeDescription()
	if transferFeeParams != nil {
		feeAmount := fee.NewTransferFeeAmount(*transferFeeParams, transfer.NewAmount(source.Currency(), requestAmount))
		chain.
			Debit(feeAmount).
			From(source).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					AccountId:                &sourceAccount.ID,
					Description:              &transferFeeDescription,
					Amount:                   pointer.ToDecimal(action.Amount().Neg()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(sourceAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(sourceAccount.Balance),
					Type:                     pointer.ToString("fee"),
					Purpose:                  pointer.ToString(constants.PurposeFeeTransfer.String()),
				}
				b.appendTransaction(transaction)

				details[constants.PurposeFeeTransfer] = &types.Detail{
					Purpose:      constants.PurposeFeeTransfer,
					Amount:       action.Amount().Neg(),
					CurrencyCode: currency.Code(),
					Transaction:  transaction,
					AccountId:    pointer.ToUint64(sourceAccount.ID),
					Account:      sourceAccount,
				}
				return err
			}).
			As("transferFee")
	}

	creditPurpose := b.creditPurpose()
	if sourceCurrency.Code() != destinationCurrency.Code() {
		rateSource := exchange.NewDirectRateSource()
		_ = rateSource.Set(exchange.NewRate(request.RateBaseCurrencyCode(), request.RateReferenceCurrencyCode(), *request.Rate))
		chain.
			Exchange(remainder).
			Using(rateSource).
			ToCurrency(destinationCurrency).
			As("destinationAmount").
			CreditFromAlias("destinationAmount").
			To(destination)
	} else {
		chain.
			Credit(remainder).
			To(destination)
	}
	chain.WithCallback(func(action transfer.Action) error {
		err := action.Perform()
		currency := action.Currency()
		transaction := &txModel.Transaction{
			RequestId:                request.Id,
			AccountId:                &destinationAccount.ID,
			Description:              &incomingDescription,
			Amount:                   pointer.ToDecimal(action.Amount()),
			IsVisible:                pointer.ToBool(true),
			AvailableBalanceSnapshot: pointer.ToDecimal(destinationAccount.AvailableAmount),
			CurrentBalanceSnapshot:   pointer.ToDecimal(destinationAccount.Balance),
			Type:                     pointer.ToString("account"),
			Purpose:                  pointer.ToString(creditPurpose.String()),
		}
		b.appendTransaction(transaction)

		details[creditPurpose] = &types.Detail{
			Purpose:      creditPurpose,
			Amount:       action.Amount(),
			CurrencyCode: currency.Code(),
			Transaction:  transaction,
			AccountId:    pointer.ToUint64(destinationAccount.ID),
			Account:      destinationAccount,
		}
		return err
	})

	if transferFeeParams != nil {
		revenuePurpose := b.transferFeePurpose()
		chain.
			CreditFromAlias("transferFee").
			To(revenue).
			WithCallback(func(action transfer.Action) error {
				err := action.Perform()
				currency := action.Currency()
				transaction := &txModel.Transaction{
					RequestId:                request.Id,
					RevenueAccountId:         &revenueAccount.ID,
					Description:              &transferFeeDescription,
					Amount:                   pointer.ToDecimal(action.Amount()),
					IsVisible:                pointer.ToBool(true),
					AvailableBalanceSnapshot: pointer.ToDecimal(revenueAccount.AvailableAmount),
					CurrentBalanceSnapshot:   pointer.ToDecimal(revenueAccount.Balance),
					Type:                     pointer.ToString("revenue"),
					Purpose:                  pointer.ToString(revenuePurpose.String()),
				}
				b.appendTransaction(transaction)

				details[revenuePurpose] = &types.Detail{
					Purpose:          revenuePurpose,
					Amount:           action.Amount(),
					CurrencyCode:     currency.Code(),
					Transaction:      transaction,
					RevenueAccountId: &revenueAccount.ID,
					RevenueAccount:   revenueAccount,
				}
				return err
			})
	}

	if exchangeMarginMultiplier.GreaterThan(decimal.NewFromInt(0)) {
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
				b.appendTransaction(transaction)

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
	err = chain.Execute()
	return details, err
}

func (b *BetweenAccounts) transferFeeDescription() string {
	return fmt.Sprintf("Transfer Fee: %s Fee", strings.ToUpper(b.subject))
}

func (b *BetweenAccounts) debitPurpose() constants.Purpose {
	return constants.Purpose(strings.ToLower(b.subject) + "_outgoing")
}

func (b *BetweenAccounts) transferFeePurpose() constants.Purpose {
	return constants.Purpose("revenue_" + strings.ToLower(b.subject) + "_transfer")
}

func (b *BetweenAccounts) creditPurpose() constants.Purpose {
	return constants.Purpose(strings.ToLower(b.subject) + "_incoming")
}
