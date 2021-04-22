package request

import (
	"fmt"

	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/Confialink/wallet-users/rpc/proto/users"

	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/olebedev/emitter"
	errorsPkg "github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountEvent "github.com/Confialink/wallet-accounts/internal/modules/account/event"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	accountService "github.com/Confialink/wallet-accounts/internal/modules/account/service"
	auth "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	cardRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/modules/fee"
	feeModel "github.com/Confialink/wallet-accounts/internal/modules/fee/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/event"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request/transfers"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	userHelper "github.com/Confialink/wallet-accounts/internal/modules/user"
	userService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/Confialink/wallet-accounts/internal/transfer"
	transferFee "github.com/Confialink/wallet-accounts/internal/transfer/fee"
)

type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

const (
	errFeeNotFound = Error(errcodes.CodeFeeParamsNotFound)
)

type Creator struct {
	db                       *gorm.DB
	userService              *userService.UserService
	currencyService          service.CurrenciesServiceInterface
	currencyProvider         transfer.CurrencyProvider
	transferFeeService       *fee.ServiceTransferFee
	requestRepository        repository.RequestRepositoryInterface
	accountsRepository       *accountRepository.AccountRepository
	cardsRepository          cardRepository.CardRepositoryInterface
	dataOwtRepository        *repository.DataOwt
	templateRepository       *repository.Template
	revenueAccountService    *accountService.RevenueAccountService
	revenueAccountRepository *accountRepository.RevenueAccountRepository
	emitter                  *emitter.Emitter
	settings                 *settings.Service
	pf                       transfers.PermissionFactory
	logger                   log15.Logger
}

func NewCreator(
	db *gorm.DB,
	uService *userService.UserService,
	currencyService service.CurrenciesServiceInterface,
	currencyProvider transfer.CurrencyProvider,
	transferFeeService *fee.ServiceTransferFee,
	requestRepository repository.RequestRepositoryInterface,
	accountsRepository *accountRepository.AccountRepository,
	cardsRepository cardRepository.CardRepositoryInterface,
	dataOwtRepository *repository.DataOwt,
	templateRepository *repository.Template,
	revenueAccountService *accountService.RevenueAccountService,
	revenueAccountRepository *accountRepository.RevenueAccountRepository,
	emitter *emitter.Emitter,
	settings *settings.Service,
	pf transfers.PermissionFactory,
	logger log15.Logger,
) *Creator {
	return &Creator{
		db:                       db,
		userService:              uService,
		currencyService:          currencyService,
		currencyProvider:         currencyProvider,
		transferFeeService:       transferFeeService,
		requestRepository:        requestRepository,
		accountsRepository:       accountsRepository,
		cardsRepository:          cardsRepository,
		templateRepository:       templateRepository,
		dataOwtRepository:        dataOwtRepository,
		revenueAccountService:    revenueAccountService,
		revenueAccountRepository: revenueAccountRepository,
		emitter:                  emitter,
		settings:                 settings,
		pf:                       pf,
		logger:                   logger,
	}
}

func (c *Creator) CreateTBARequest(form *form.TBA, user *users.User, db *gorm.DB) (request *model.Request, err error) {
	accountFrom, err := getAccountWithTypeForUpdateById(db, *form.AccountIdFrom)
	logger := c.logger.New("action", "CreateTBARequest")
	if err != nil {
		return
	}

	accountTo, err := getAccountWithTypeForUpdateById(db, *form.AccountIdTo)
	if err != nil {
		return
	}

	// Create if not exist TODO: refactor - revenue accounts must exist in advance
	revenueAccount, err := c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountFrom.Type.CurrencyCode, db)
	if err != nil {
		logger.Error("failed to find or create revenue account", "error", err)
		return
	}

	revenueAccount, err = getRevenueAccountForUpdateById(db, revenueAccount.ID)
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(accountFrom.Type.CurrencyCode, accountTo.Type.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyIdFrom", accountFrom.Type.CurrencyCode, "currencyIdTo", accountTo.Type.CurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "TBA", nil)
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectTransferBetweenAccounts
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountTo.Type.CurrencyCode,
		Amount:                &amount,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate.Rate,
	}

	shouldExecute, err := c.shouldExecute(request)
	if err != nil {
		return
	}
	request.IsVisible = pointer.ToBool(!shouldExecute)

	//TODO: refactor
	requestInput := request.GetInput()
	requestInput.Set("transferFeeParams", params)
	requestInput.Set("sourceAccountId", int64(*form.AccountIdFrom))
	requestInput.Set("destinationAccountId", int64(*form.AccountIdTo))
	requestInput.Set("sourceAccountNumber", accountFrom.Number)
	requestInput.Set("destinationAccountNumber", accountTo.Number)
	requestInput.Set("revenueAccountId", int64(revenueAccount.ID))
	requestInput.Set("exchangeMarginPercent", rate.ExchangeMargin)

	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return
	}

	input := transfers.NewBetweenAccountsInput(
		accountFrom,
		accountTo,
		revenueAccount,
		rate.ExchangeMargin,
		params,
	)
	tba := transfers.NewBetweenAccounts("TBA", c.currencyProvider, input, db, c.pf)
	if shouldExecute {
		details, err := tba.Execute(request)
		if err == nil {
			<-c.emitter.Emit(
				event.RequestExecuted,
				&event.ContextRequestExecuted{
					Tx:      db,
					Request: request,
					Details: details,
				},
			)
			accountEvent.TriggerBalanceChanged(c.emitter, db, *request.Subject, details)
		}
		return request, err
	}

	details, err := tba.Pending(request)

	if err == nil {
		eventContext := &event.ContextRequestPending{
			Tx:      db,
			Request: request,
			Details: details,
		}
		<-c.emitter.Emit(event.RequestPendingApproval, eventContext)
	}

	return
}

func (c *Creator) EvaluateTBARequest(form *form.TBAPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateTBARequest")
	accountFrom, err := c.accountsRepository.FindByID(*form.AccountIdFrom)
	if err != nil {
		return
	}

	accountTo, err := c.accountsRepository.FindByID(*form.AccountIdTo)
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(accountFrom.Type.CurrencyCode, accountTo.Type.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyIdFrom", accountFrom.Type.CurrencyCode, "currencyIdTo", accountTo.Type.CurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	subject := constants.SubjectTransferBetweenAccounts
	request := &model.Request{
		Amount:                &amount,
		Subject:               &subject,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate.Rate,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountTo.Type.CurrencyCode,
		IsInitiatedBySystem:   pointer.ToBool(false),
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "TBA", nil)
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}

	input := transfers.NewBetweenAccountsInput(
		accountFrom,
		accountTo,
		stubRevenueAccount(accountFrom.Type.CurrencyCode),
		rate.ExchangeMargin,
		params,
	)

	tba := transfers.NewBetweenAccounts("TBA", c.currencyProvider, input, c.db, c.pf)

	return tba.Evaluate(request)
}

func (c *Creator) CreateTBURequest(form *form.TBU, user *users.User, db *gorm.DB) (request *model.Request, err error) {
	logger := c.logger.New("action", "CreateTBURequest")

	accountFrom, err := getAccountWithTypeForUpdateById(db, *form.AccountIdFrom)
	if err != nil {
		logger.Error("failed to find source account", "error", err, "accountId", *form.AccountIdFrom)
		return
	}

	accountTo, err := getAccountWithTypeForUpdateByNumber(db, *form.AccountNumberTo)
	if err != nil {
		logger.Error("failed to find destination account", "error", err, "accountNumber", *form.AccountNumberTo)
		return
	}

	// TODO: have updated status in the service directly
	if user.RoleName == "client" {
		usersFull, _ := c.userService.GetFullByUIDs([]string{accountTo.UserId}, []string{"UID", "Username", "IsCorporate", "Status"})
		if usersFull[0].Status != "active" {
			err = errcodes.CreatePublicError(errcodes.CodeUserInvalidStatus, "The “CreditFromAlias to” user has invalid status.")
			return
		}
	}

	revenueAccount, err := c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountFrom.Type.CurrencyCode, db)
	if err != nil {
		logger.Error("failed to find or create revenue account", "error", err)
		return
	}

	revenueAccount, err = getRevenueAccountForUpdateById(db, revenueAccount.ID)
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(accountFrom.Type.CurrencyCode, accountTo.Type.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyIdFrom", accountFrom.Type.CurrencyCode, "currencyIdTo", accountTo.Type.CurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "TBU", nil)
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectTransferBetweenUsers
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountTo.Type.CurrencyCode,
		Amount:                &amount,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate.Rate,
	}

	shouldExecute, err := c.shouldExecute(request)
	if err != nil {
		return
	}
	request.IsVisible = pointer.ToBool(!shouldExecute)

	requestInput := request.GetInput()
	requestInput.Set("transferFeeParams", params)
	requestInput.Set("sourceAccountId", *form.AccountIdFrom)
	requestInput.Set("destinationAccountId", accountTo.ID)
	requestInput.Set("sourceAccountNumber", accountFrom.Number)
	requestInput.Set("destinationAccountNumber", accountTo.Number)
	requestInput.Set("revenueAccountId", revenueAccount.ID)
	requestInput.Set("exchangeMarginPercent", rate.ExchangeMargin)

	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return
	}

	if !isSystem {
		err = c.saveTemplateIfRequired(db, user, subject, form)
		if err != nil {
			return
		}
	}

	input := transfers.NewBetweenAccountsInput(
		accountFrom,
		accountTo,
		revenueAccount,
		rate.ExchangeMargin,
		params,
	)

	tbu := transfers.NewBetweenAccounts("TBU", c.currencyProvider, input, db, c.pf)

	if shouldExecute {
		details, err := tbu.Execute(request)
		if err == nil {
			<-c.emitter.Emit(
				event.RequestExecuted,
				&event.ContextRequestExecuted{
					Tx:      db,
					Request: request,
					Details: details,
				},
			)
			accountEvent.TriggerBalanceChanged(c.emitter, db, *request.Subject, details)
		}
		return request, err
	}

	details, err := tbu.Pending(request)

	if err == nil {
		eventContext := &event.ContextRequestPending{
			Tx:      db,
			Request: request,
			Details: details,
		}
		<-c.emitter.Emit(event.RequestPendingApproval, eventContext)
	}
	return
}

func (c *Creator) EvaluateTBURequest(form *form.TBUPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateTBURequest")
	accountFrom, err := c.accountsRepository.FindByID(*form.AccountIdFrom)
	if err != nil {
		return
	}

	accountTo, err := c.accountsRepository.FindByNumber(*form.AccountNumberTo)
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(accountFrom.Type.CurrencyCode, accountTo.Type.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyIdFrom", accountFrom.Type.CurrencyCode, "currencyIdTo", accountTo.Type.CurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	subject := constants.SubjectTransferBetweenUsers
	request := &model.Request{
		Amount:                &amount,
		Subject:               &subject,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate.Rate,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountTo.Type.CurrencyCode,
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "TBU", nil)
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}
	input := transfers.NewBetweenAccountsInput(
		accountFrom,
		accountTo,
		stubRevenueAccount(accountFrom.Type.CurrencyCode),
		rate.ExchangeMargin,
		params,
	)

	tbu := transfers.NewBetweenAccounts("TBU", c.currencyProvider, input, c.db, c.pf)
	return tbu.Evaluate(request)
}

func (c *Creator) CreateOWTRequest(form *form.OWT, user *users.User, db *gorm.DB) (request *model.Request, err error) {
	accountFrom, err := getAccountWithTypeForUpdateById(db, *form.AccountIdFrom)
	logger := c.logger.New("action", "CreateOWTRequest")
	if err != nil {
		logger.Error("failed to find source account", "error", err, "accountId", *form.AccountIdFrom)
		return
	}

	revenueAccount, err := c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountFrom.Type.CurrencyCode, db)
	if err != nil {
		logger.Error("failed to find or create revenue account", "error", err)
		return
	}

	//!for update
	revenueAccount, err = getRevenueAccountForUpdateById(db, revenueAccount.ID)
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(*form.ReferenceCurrencyCode, accountFrom.Type.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyIdFrom", accountFrom.Type.CurrencyCode, "currencyIdTo", form.ReferenceCurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}
	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "OWT", form.FeeId)
	if form.FeeId != nil && errorsPkg.Cause(err) == errFeeNotFound {
		return nil, errcodes.CreatePublicError(errcodes.CodeFeeParamsNotFound)
	}
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectTransferOutgoingWireTransfer
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: form.ReferenceCurrencyCode,
		InputAmount:           &amount,
		RateDesignation:       model.RateDesignationReferenceBase,
		Rate:                  &rate.Rate,
		IsVisible:             pointer.ToBool(true),
	}
	requestInput := request.GetInput()
	requestInput.Set("transferFeeParams", params)
	requestInput.Set("sourceAccountId", *form.AccountIdFrom)
	requestInput.Set("sourceAccountNumber", accountFrom.Number)
	requestInput.Set("revenueAccountId", revenueAccount.ID)
	requestInput.Set("exchangeMarginPercent", rate.ExchangeMargin)
	requestInput.Set("refMessage", *form.RefMessage)
	requestInput.Set("beneficiaryCustomerAccountName", *form.CustomerName)

	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return
	}

	if !isAdmin && !isSystem {
		err = c.saveTemplateIfRequired(db, user, subject, form)
		if err != nil {
			return
		}
	}

	data := form.NewDataOwt()
	data.RequestId = request.Id
	dataRepoTx := c.dataOwtRepository.WrapContext(db)
	err = dataRepoTx.Create(data)
	if err != nil {
		return
	}

	input := transfers.NewOwtInput(
		accountFrom,
		revenueAccount,
		rate.ExchangeMargin,
		params,
		*form.CustomerName,
		*form.RefMessage,
	)

	owt := transfers.NewOutgoingWireTransfer(input, c.currencyProvider, db, c.pf)

	details, err := owt.Pending(request)
	if err == nil {
		eventContext := &event.ContextRequestPending{
			Tx:      db,
			Request: request,
			Details: details,
		}
		<-c.emitter.Emit(event.RequestPendingApproval, eventContext)
	}

	return
}

func (c *Creator) EvaluateOWTRequest(form *form.OWTPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateOWTRequest")
	accountFrom, err := c.accountsRepository.FindByID(*form.AccountIdFrom)
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(*form.ReferenceCurrencyCode, accountFrom.Type.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyIdFrom", accountFrom.Type.CurrencyCode, "currencyIdTo", *form.ReferenceCurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	subject := constants.SubjectTransferOutgoingWireTransfer
	request := &model.Request{
		InputAmount:           &amount,
		Subject:               &subject,
		RateDesignation:       model.RateDesignationReferenceBase,
		Rate:                  &rate.Rate,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: form.ReferenceCurrencyCode,
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "OWT", form.FeeId)
	if form.FeeId != nil && errorsPkg.Cause(err) == errFeeNotFound {
		return nil, errcodes.CreatePublicError(errcodes.CodeFeeParamsNotFound)
	}
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}

	input := transfers.NewOwtInput(
		accountFrom,
		stubRevenueAccount(accountFrom.Type.CurrencyCode),
		rate.ExchangeMargin,
		params,
		"",
		"",
	)

	owt := transfers.NewOutgoingWireTransfer(input, c.currencyProvider, c.db, c.pf)
	return owt.Evaluate(request)
}

func (c *Creator) CreateCFTRequest(form *form.CFT, user *users.User, db *gorm.DB) (request *model.Request, err error) {
	logger := c.logger.New("action", "CreateCFTRequest")

	card, err := getCardWithTypeForUpdateById(db, *form.CardIdTo)
	if err != nil {
		return
	}

	accountFrom, err := getAccountWithTypeForUpdateById(db, *form.AccountIdFrom)
	if err != nil {
		logger.Error("failed to find source account", "error", err, "accountId", *form.AccountIdFrom)
		return
	}

	rate, err := c.getRateForCurrencies(accountFrom.Type.CurrencyCode, *card.CardType.CurrencyCode)

	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyCodeFrom", accountFrom.Type.CurrencyCode, "currencyCodeTo", *card.CardType.CurrencyCode)
		return
	}

	revenueAccount, err := c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountFrom.Type.CurrencyCode, db)
	if err != nil {
		logger.Error("failed to find or create revenue account", "error", err)
		return
	}

	revenueAccount, err = getRevenueAccountForUpdateById(db, revenueAccount.ID)
	if err != nil {
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "CFT", nil)
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectCardFundingTransfer
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: card.CardType.CurrencyCode,
		Amount:                &amount,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate.Rate,
		IsVisible:             pointer.ToBool(true),
	}
	requestInput := request.GetInput()
	requestInput.Set("transferFeeParams", params)
	requestInput.Set("sourceAccountId", *form.AccountIdFrom)
	requestInput.Set("sourceAccountNumber", accountFrom.Number)
	requestInput.Set("destinationCardId", *form.CardIdTo)
	requestInput.Set("destinationCardNumber", card.Number)
	requestInput.Set("revenueAccountId", revenueAccount.ID)
	requestInput.Set("exchangeMarginPercent", rate.ExchangeMargin)

	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return
	}

	if !isAdmin && !isSystem {
		err = c.saveTemplateIfRequired(db, user, subject, form)
		if err != nil {
			return
		}
	}

	input := transfers.NewCFTInput(
		accountFrom,
		card,
		revenueAccount,
		rate.ExchangeMargin,
		params,
	)

	cft := transfers.NewCardFunding(c.currencyProvider, input, db, c.pf)
	details, err := cft.Pending(request)
	if err == nil {
		eventContext := &event.ContextRequestPending{
			Tx:      db,
			Request: request,
			Details: details,
		}
		<-c.emitter.Emit(event.RequestPendingApproval, eventContext)
	}

	return
}

func (c *Creator) EvaluateCFTRequest(form *form.CFTPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateCFTRequest")
	accountFrom, err := c.accountsRepository.FindByID(*form.AccountIdFrom)
	if err != nil {
		return
	}

	card, err := c.cardsRepository.Get(*form.CardIdTo, list_params.NewIncludes("include=CardType"))
	if err != nil {
		return
	}

	rate, err := c.getRateForCurrencies(accountFrom.Type.CurrencyCode, *card.CardType.CurrencyCode)
	if err != nil {
		logger.Error("failed to obtain rate", "error", err, "currencyCodeFrom", accountFrom.Type.CurrencyCode, "currencyCodeTo", card.CardType.CurrencyCode)
		return
	}

	amount, err := decimal.NewFromString(*form.OutgoingAmount)
	if err != nil {
		return
	}

	subject := constants.SubjectCardFundingTransfer
	request := &model.Request{
		Amount:                &amount,
		Subject:               &subject,
		Rate:                  &rate.Rate,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: card.CardType.CurrencyCode,
	}

	params, err := c.getFeeParams(c.db, accountFrom.UserId, accountFrom.Type.CurrencyCode, "CFT", nil)
	if err != nil && errorsPkg.Cause(err) != errFeeNotFound {
		return
	}
	input := transfers.NewCFTInput(
		accountFrom,
		card,
		stubRevenueAccount(accountFrom.Type.CurrencyCode),
		rate.ExchangeMargin,
		params,
	)

	cft := transfers.NewCardFunding(c.currencyProvider, input, c.db, c.pf)

	details, err = cft.Evaluate(request)
	return
}

func (c *Creator) EvaluateCARequest(form *form.CAPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateCARequest")

	accountTo, err := c.accountsRepository.FindByID(form.AccountId)
	if err != nil {
		logger.Error("failed to find destination account", "error", err, "accountId", form.AccountId)
		return
	}

	rate := decimal.NewFromFloat(1)
	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		return
	}

	subject := constants.SubjectCreditAccount
	request := &model.Request{
		Amount:                &amount,
		Subject:               &subject,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate,
		BaseCurrencyCode:      &accountTo.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountTo.Type.CurrencyCode,
	}
	var feeParams *transferFee.TransferFeeParams
	if *form.ApplyIwtFee {
		feeParams, err = c.getFeeParams(c.db, accountTo.UserId, *request.BaseCurrencyCode, "IWT", nil)
		if err != nil {
			if errorsPkg.Cause(err) == errFeeNotFound {
				return nil, errcodes.CreatePublicError(
					errcodes.CodeFeeParamsNotFound,
					"The IWT fee is not specified for the user group, this user belongs to.",
				)
			}
			return
		}
	}
	// Create if not exist
	revenueAccount, err := c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountTo.Type.CurrencyCode, c.db)
	if err != nil {
		logger.Error("failed to find or create revenue account", "error", err)
		return
	}
	caInput := transfers.NewCreditAccountInput(
		accountTo,
		*form.ApplyIwtFee,
		*form.DebitFromRevenueAccount,
		revenueAccount,
		feeParams,
	)

	ca := transfers.NewCreditAccount(c.db, caInput, c.currencyProvider)

	return ca.Evaluate(request)
}

func (c *Creator) CreateCARequest(
	form *form.CA,
	user *users.User,
	db *gorm.DB,
	isInitialBalanceRequest ...bool,
) (request *model.Request, err error) {
	logger := c.logger.New("action", "CreateCARequest")

	accountTo, err := getAccountWithTypeForUpdateById(db, form.AccountId)
	if err != nil {
		logger.Error("failed to find destination account", "error", err, "accountId", form.AccountId)
		return
	}

	rate := decimal.NewFromFloat(1)

	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectCreditAccount
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           &form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &accountTo.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountTo.Type.CurrencyCode,
		Amount:                &amount,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate,
		IsVisible:             pointer.ToBool(false),
	}
	requestInput := request.GetInput()
	requestInput.Set("debitFromRevenueAccount", *form.DebitFromRevenueAccount)
	requestInput.Set("destinationAccountId", form.AccountId)
	requestInput.Set("applyIwtFee", *form.ApplyIwtFee)
	if len(isInitialBalanceRequest) > 0 && isInitialBalanceRequest[0] {
		requestInput.Set("isInitialBalanceRequest", true)
	}

	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return
	}

	// Create if not exist
	revenueAccount, err := c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountTo.Type.CurrencyCode, db)
	if err != nil {
		logger.Error("failed to find or create revenue account", "error", err)
		return
	}

	revenueAccount, err = getRevenueAccountForUpdateById(db, revenueAccount.ID)
	if err != nil {
		return
	}

	var feeParams *transferFee.TransferFeeParams
	if *form.ApplyIwtFee {
		feeParams, err = c.getFeeParams(db, accountTo.UserId, *request.BaseCurrencyCode, "IWT", nil)
		if err != nil {
			if errorsPkg.Cause(err) == errFeeNotFound {
				return nil, errcodes.CreatePublicError(
					errcodes.CodeFeeParamsNotFound,
					"The IWT fee is not specified for the user group, this user belongs to.",
				)
			}
			return
		}
	}
	caInput := transfers.NewCreditAccountInput(
		accountTo,
		*form.ApplyIwtFee,
		*form.DebitFromRevenueAccount,
		revenueAccount,
		feeParams,
	)

	unit := transfers.NewCreditAccount(db, caInput, c.currencyProvider)
	details, err := unit.Execute(request)
	if err == nil {
		<-c.emitter.Emit(
			event.RequestExecuted,
			&event.ContextRequestExecuted{
				Tx:      db,
				Request: request,
				Details: details,
			},
		)
		accountEvent.TriggerBalanceChanged(c.emitter, db, *request.Subject, details)
	}
	return
}

func (c *Creator) EvaluateDARequest(form *form.DAPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateDARequest")

	accountFrom, err := c.accountsRepository.FindByID(form.AccountId)
	if err != nil {
		logger.Error("failed to find source account", "error", err, "accountId", form.AccountId)
		return
	}

	rate := decimal.NewFromFloat(1)
	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		return
	}

	subject := constants.SubjectDebitAccount
	request := &model.Request{
		Amount:                &amount,
		Subject:               &subject,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountFrom.Type.CurrencyCode,
	}

	input := transfers.NewDaInput(
		accountFrom,
		stubRevenueAccount(accountFrom.Type.CurrencyCode),
		*form.CreditToRevenueAccount,
		false,
	)

	da := transfers.NewDebitAccount(c.db, input, c.currencyProvider)
	return da.Evaluate(request)
}

func (c *Creator) CreateDARequest(form *form.DA, user *users.User, db *gorm.DB) (request *model.Request, err error) {
	logger := c.logger.New("action", "CreateDARequest")

	rate := decimal.NewFromFloat(1)

	accountFrom, err := getAccountWithTypeForUpdateById(db, form.AccountId)
	if err != nil {
		logger.Error("failed to find source account", "error", err, "accountId", form.AccountId)
		return
	}

	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectDebitAccount
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           &form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &accountFrom.Type.CurrencyCode,
		ReferenceCurrencyCode: &accountFrom.Type.CurrencyCode,
		Amount:                &amount,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  &rate,
		IsVisible:             pointer.ToBool(false),
	}
	request.GetInput().Set("creditToRevenueAccount", *form.CreditToRevenueAccount)
	request.GetInput().Set("sourceAccountId", form.AccountId)
	request.GetInput().Set("sourceAccountNumber", accountFrom.Number)
	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return
	}
	var revenueAccount *accountModel.RevenueAccountModel
	if form.CreditToRevenueAccount != nil && *form.CreditToRevenueAccount {
		revenueAccount, err = c.revenueAccountService.FindOrCreateDefaultByCurrencyCode(accountFrom.Type.CurrencyCode, db)
		if err != nil {
			logger.Error("failed to find or create revenue account", "error", err)
			return
		}
		revenueAccount, err = getRevenueAccountForUpdateById(db, revenueAccount.ID)
		if err != nil {
			return
		}
	}
	input := transfers.NewDaInput(
		accountFrom,
		revenueAccount,
		*form.CreditToRevenueAccount,
		true,
	)

	da := transfers.NewDebitAccount(db, input, c.currencyProvider)
	details, err := da.Execute(request)
	if err == nil {
		<-c.emitter.Emit(
			event.RequestExecuted,
			&event.ContextRequestExecuted{
				Tx:      db,
				Request: request,
				Details: details,
			},
		)
		accountEvent.TriggerBalanceChanged(c.emitter, db, *request.Subject, details)
	}
	return
}

func (c *Creator) EvaluateDRARequest(form *form.DRAPreview, user *users.User) (details types.Details, err error) {
	logger := c.logger.New("action", "EvaluateDRRequest")

	account, err := c.revenueAccountRepository.FindByID(form.RevenueAccountId)
	if err != nil {
		logger.Error("failed to find revenue account", "error", err, "accountId", form.RevenueAccountId)
		return
	}

	rate := decimal.NewFromFloat(1)

	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		return
	}

	subject := constants.SubjectDebitRevenueAccount
	request := &model.Request{
		Amount:           &amount,
		Subject:          &subject,
		RateDesignation:  model.RateDesignationBaseReference,
		Rate:             &rate,
		BaseCurrencyCode: &account.CurrencyCode,
	}

	input := transfers.NewDraInput(account)
	dra := transfers.NewDeductRevenueAccount(c.db, input, c.currencyProvider)

	return dra.Evaluate(request)
}

func (c *Creator) CreateDRARequest(form *form.DRA, user *users.User, db *gorm.DB) (request *model.Request, err error) {
	logger := c.logger.New("action", "CreateDARequest")

	account, err := getRevenueAccountForUpdateById(db, form.RevenueAccountId)

	if err != nil {
		logger.Error("failed to find revenue account", "error", err, "accountId", form.RevenueAccountId)
		return
	}

	amount, err := decimal.NewFromString(form.Amount)
	if err != nil {
		return
	}

	isAdmin, isSystem := c.GetIsAdminIsSystem(user)
	subject := constants.SubjectDebitRevenueAccount
	status := constants.StatusNew
	request = &model.Request{
		Subject:               &subject,
		Description:           form.Description,
		Status:                &status,
		UserId:                &user.UID,
		IsInitiatedByAdmin:    &isAdmin,
		IsInitiatedBySystem:   &isSystem,
		BaseCurrencyCode:      &account.CurrencyCode,
		ReferenceCurrencyCode: &account.CurrencyCode,
		RateDesignation:       model.RateDesignationBaseReference,
		Rate:                  pointer.ToDecimal(decimal.NewFromInt(1)),
		Amount:                &amount,
		IsVisible:             pointer.ToBool(false),
	}
	requestInput := request.GetInput()
	requestInput.Set("revenueAccountId", account.ID)

	reqRepoTx := c.requestRepository.WrapContext(db)

	err = reqRepoTx.Create(request)
	if err != nil {
		return nil, err
	}

	input := transfers.NewDraInput(account)
	dra := transfers.NewDeductRevenueAccount(db, input, c.currencyProvider)

	details, err := dra.Execute(request)
	if err == nil {
		<-c.emitter.Emit(
			event.RequestExecuted,
			&event.ContextRequestExecuted{
				Tx:      db,
				Request: request,
				Details: details,
			},
		)
		accountEvent.TriggerBalanceChanged(c.emitter, db, *request.Subject, details)
	}
	return
}

func (c *Creator) GetIsAdminIsSystem(user *users.User) (bool, bool) {
	if userHelper.IsSystemUser(user) {
		return false, true
	}

	return user.RoleName == auth.RoleAdmin || user.RoleName == auth.RoleRoot, false
}

func (c *Creator) saveTemplateIfRequired(db *gorm.DB, user *users.User, subject constants.Subject, template interface{}) error {
	if tmp, ok := template.(form.Template); ok {
		if !tmp.SaveAsTemplate() {
			return nil
		}
		templateModel := &model.Template{
			Name:           pointer.ToString(tmp.TemplateName()),
			RequestSubject: &subject,
			UserId:         &user.UID,
		}
		_ = templateModel.SetData(tmp.TemplateData())
		return c.templateRepository.WrapContext(db).Create(templateModel)
	}
	return nil
}

func (c *Creator) getRateForCurrencies(currencyCodeFrom, currencyCodeTo string) (*service.Rate, error) {
	rate := &service.Rate{
		Rate:           decimal.NewFromInt(1),
		ExchangeMargin: decimal.NewFromInt(0),
	}
	if currencyCodeFrom != currencyCodeTo {
		var err error
		rate, err = c.currencyService.GetCurrenciesRateByCodes(currencyCodeFrom, currencyCodeTo)
		if err != nil {
			return rate, errorsPkg.Wrapf(
				err,
				"failed get rate for currencies %s/%s",
				currencyCodeFrom,
				currencyCodeTo,
			)
		}
		if rate.Rate.LessThanOrEqual(decimal.Zero) {
			return rate, errcodes.CreatePublicError(errcodes.CodeInvalidExchangeRate)
		}
	}
	return rate, nil
}

func (c *Creator) getFeeParams(db *gorm.DB, userUID, currencyCode, subject string, feeId *uint64) (*transferFee.TransferFeeParams, error) {
	logger := c.logger.New("method", "getFeeParams")
	accountOwner, err := c.userService.GetByUID(userUID)
	if err != nil {
		logger.Error(
			"failed to fetch account owner",
			"error", err,
			"userUID", userUID,
		)
	}
	paramsQuery := &fee.ParamsQuery{
		UserGroupId:    accountOwner.GroupId,
		CurrencyCode:   currencyCode,
		RequestSubject: constants.Subject(subject),
		FeeId:          feeId,
	}
	params, err := c.transferFeeService.FindParams(paramsQuery, db)

	if err != nil {
		return nil, errFeeNotFound
	}
	return transferFeeModelToParams(params), nil
}

func (c *Creator) adminApprovalRequired(subject string) (bool, error) {
	actionRequiredSettingName := settings.Name(fmt.Sprintf("%s_action_required", subject))
	return c.settings.Bool(actionRequiredSettingName)
}

func (c *Creator) shouldExecute(request *model.Request, adminApprovalRequired ...bool) (bool, error) {
	if *request.IsInitiatedBySystem || *request.IsInitiatedByAdmin {
		return true, nil
	}
	approvalRequired, err := c.adminApprovalRequired(request.Subject.String())
	return !approvalRequired, err
}

func transferFeeModelToParams(feeModelParams *feeModel.TransferFeeParameters) *transferFee.TransferFeeParams {
	result := &transferFee.TransferFeeParams{}
	if feeModelParams.Base != nil {
		result.Base = *feeModelParams.Base
	}
	if feeModelParams.Percent != nil {
		result.Percent = *feeModelParams.Percent
	}
	if feeModelParams.Min != nil {
		result.Min = *feeModelParams.Min
	}
	if feeModelParams.Max != nil {
		result.Max = *feeModelParams.Max
	}
	return result
}

func getAccountWithTypeForUpdateById(db *gorm.DB, accountId uint64) (*accountModel.Account, error) {
	account := &accountModel.Account{}
	err := db.
		Preload("Type").
		Raw("SELECT * FROM `accounts` WHERE `accounts`.`id` = ? FOR UPDATE", accountId).
		Find(account).
		Error
	return account, err
}

func getAccountWithTypeForUpdateByNumber(db *gorm.DB, number string) (*accountModel.Account, error) {
	account := &accountModel.Account{}
	err := db.
		Preload("Type").
		Raw("SELECT * FROM `accounts` WHERE `accounts`.`number` = ? FOR UPDATE", number).
		Find(account).
		Error
	return account, err
}

func getRevenueAccountForUpdateById(db *gorm.DB, accountId uint64) (*accountModel.RevenueAccountModel, error) {
	account := &accountModel.RevenueAccountModel{}
	err := db.
		Raw("SELECT * FROM `revenue_accounts` WHERE `revenue_accounts`.`id` = ? FOR UPDATE", accountId).
		Find(account).
		Error
	return account, err
}

func getCardWithTypeForUpdateById(db *gorm.DB, cardId uint32) (*cardModel.Card, error) {
	card := &cardModel.Card{}
	err := db.
		Preload("CardType").
		Raw("SELECT * FROM `cards` WHERE `cards`.`id` = ? FOR UPDATE", cardId).
		Find(card).
		Error
	return card, err
}

func stubRevenueAccount(currencyCode string) *accountModel.RevenueAccountModel {
	return &accountModel.RevenueAccountModel{
		RevenueAccountPublic: accountModel.RevenueAccountPublic{
			CurrencyCode: currencyCode,
		},
	}
}
