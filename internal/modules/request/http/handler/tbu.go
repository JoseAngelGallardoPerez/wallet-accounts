package handler

import (
	"log"
	"net/http"

	errorsPkg "github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	currencyService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	userService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

type TbuHandler struct {
	contextService    service.ContextInterface
	accountRepository *accountRepository.AccountRepository
	requestCreator    *request.Creator
	userService       *userService.UserService
	currencyService   currencyService.CurrenciesServiceInterface
	db                *gorm.DB
	logger            log15.Logger
}

func NewTbuHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	requestCreator *request.Creator,
	userService *userService.UserService,
	currencyService currencyService.CurrenciesServiceInterface,
	db *gorm.DB,
	logger log15.Logger,

) *TbuHandler {
	return &TbuHandler{
		contextService:    contextService,
		accountRepository: accountRepository,
		requestCreator:    requestCreator,
		userService:       userService,
		currencyService:   currencyService,
		db:                db,
		logger:            logger.New("Handler", "TbuHandler"),
	}
}

func (t *TbuHandler) CreatePreviewAdmin(c *gin.Context) {
	ownerId := c.Param("userId")

	tbuForm := &form.TBUPreview{}

	if err := c.ShouldBind(tbuForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbuForm.AccountIdFrom)
	if err != nil {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: err.Error()})
		return
	}
	destinationAcc, err := t.findDestinationAccount(c, tbuForm)
	if err != nil {
		return
	}

	if sourceAcc.UserId != ownerId || destinationAcc.UserId == ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateTBURequest(tbuForm, user)

	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeTBUIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeTBUIncoming is not set"})
		return
	}

	previewData := &preview{
		Details:              details,
		IncomingAmount:       detail.Amount.String(),
		IncomingCurrencyCode: detail.CurrencyCode,
	}

	c.JSON(http.StatusOK, response.New().SetData(previewData))
}

func (t *TbuHandler) CreatePreviewUser(c *gin.Context) {
	initiator := t.contextService.MustGetCurrentUser(c)
	tbuForm := &form.TBUPreview{}

	if err := c.ShouldBind(tbuForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbuForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbuHandler unable to find account %d: %s", *tbuForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}
	destinationAcc, err := t.findDestinationAccount(c, tbuForm)
	if err != nil {
		return
	}

	if sourceAcc.UserId != initiator.UID || destinationAcc.UserId == initiator.UID {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	details, err := t.requestCreator.EvaluateTBURequest(tbuForm, initiator)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))

		return
	}

	detail, ok := details[transactionConstants.PurposeTBUIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeTBUIncoming is not set"})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{
		Details:              details,
		IncomingAmount:       detail.Amount.String(),
		IncomingCurrencyCode: detail.CurrencyCode,
	}))
}

func (t *TbuHandler) CreateRequestUser(c *gin.Context) {
	initiator := t.contextService.MustGetCurrentUser(c)
	tbuForm := &form.TBU{}

	if err := c.ShouldBind(tbuForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbuForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbuHandler unable to find account %d: %s", *tbuForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	destinationAcc, err := t.findDestinationAccount(c, tbuForm.ToTBUPreview())
	if err != nil {
		return
	}

	if sourceAcc.UserId != initiator.UID || destinationAcc.UserId == initiator.UID {
		errcodes.AddError(c, errcodes.CodeForbidden)
		return
	}

	formIncomingAmount, _ := decimal.NewFromString(*tbuForm.IncomingAmount)

	details, err := t.requestCreator.EvaluateTBURequest(tbuForm.ToTBUPreview(), initiator)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeTBUIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeTBUIncoming is not set"})
		return
	}

	if !detail.Amount.Equal(formIncomingAmount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateTBURequest(tbuForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (t *TbuHandler) CreateRequestAdmin(c *gin.Context) {
	ownerId := c.Param("userId")

	initiator := t.contextService.MustGetCurrentUser(c)
	tbuForm := &form.TBU{}

	if err := c.ShouldBind(tbuForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbuForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbuHandler unable to find account %d: %s", *tbuForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	destinationAcc, err := t.findDestinationAccount(c, tbuForm.ToTBUPreview())
	if err != nil {
		return
	}

	if sourceAcc.UserId != ownerId || destinationAcc.UserId == ownerId {
		errcodes.AddError(c, errcodes.CodeForbidden)
		return
	}

	details, err := t.requestCreator.EvaluateTBURequest(tbuForm.ToTBUPreview(), initiator)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeTBUIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeTBUIncoming is not set"})
		return
	}

	formIncomingAmount, _ := decimal.NewFromString(*tbuForm.IncomingAmount)
	if !detail.Amount.Equal(formIncomingAmount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateTBURequest(tbuForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (t *TbuHandler) findDestinationAccount(
	c *gin.Context, form *form.TBUPreview,
) (*model.Account, error) {
	account, err := t.accountRepository.FindByNumber(*form.AccountNumberTo)
	if err != nil {
		t.logger.Info("destination account is not found", "acc#", *form.AccountNumberTo, "error", err)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return nil, err
	}

	return account, nil
}

func (t *TbuHandler) Receive(c *gin.Context) {
	receiveForm := &form.TBUReceive{}

	if err := c.ShouldBind(receiveForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	account, err := t.accountRepository.FindByID(*receiveForm.AccountIdTo)
	if err != nil {
		t.logger.Info("destination account is not found", "acc#", *receiveForm.AccountIdTo, "error", err)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	res := map[string]interface{}{
		"number": account.Number,
		"qrCode": "data:image/png;base64,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",
	}

	c.JSON(http.StatusOK, response.New().SetData(res))
}
