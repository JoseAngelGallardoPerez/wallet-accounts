package handler

import (
	"net/http"

	errorsPkg "github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	currencyService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/model"
	moneyRequestService "github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	userService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

type MoneyRequestTbuHandler struct {
	contextService      service.ContextInterface
	accountRepository   *accountRepository.AccountRepository
	requestCreator      *request.Creator
	userService         *userService.UserService
	currencyService     currencyService.CurrenciesServiceInterface
	moneyRequestService *moneyRequestService.MoneyRequest
	db                  *gorm.DB
	logger              log15.Logger
}

func NewMoneyRequestTbuHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	requestCreator *request.Creator,
	userService *userService.UserService,
	currencyService currencyService.CurrenciesServiceInterface,
	moneyRequestService *moneyRequestService.MoneyRequest,
	db *gorm.DB,
	logger log15.Logger,

) *MoneyRequestTbuHandler {
	return &MoneyRequestTbuHandler{
		contextService:      contextService,
		accountRepository:   accountRepository,
		requestCreator:      requestCreator,
		userService:         userService,
		currencyService:     currencyService,
		moneyRequestService: moneyRequestService,
		db:                  db,
		logger:              logger.New("Handler", "TbuHandler"),
	}
}

func (t *MoneyRequestTbuHandler) CreatePreviewUser(c *gin.Context) {
	initiator := t.contextService.MustGetCurrentUser(c)
	requestForm := &form.MoneyRequestTBUPreview{}

	if err := c.ShouldBind(requestForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	moneyRequest, typedErr := t.moneyRequestService.GetByTargetUID(*requestForm.MoneyRequestId, initiator.UID)
	if typedErr != nil {
		errorsPkg.AddShouldBindError(c, typedErr)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*requestForm.AccountIdFrom)
	if err != nil {
		t.logger.Info("tbuHandler unable to find account %d: %s", *requestForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	destinationAcc, err := t.accountRepository.FindByID(moneyRequest.RecipientAccountID)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		t.logger.Info("tbuHandler unable to find account %d: %s", moneyRequest.RecipientAccountID, err.Error())
		return
	}

	if sourceAcc.UserId != initiator.UID || destinationAcc.UserId == initiator.UID {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	tbuForm := form.TBUPreview{
		AccountIdFrom:   requestForm.AccountIdFrom,
		AccountNumberTo: &destinationAcc.Number,
		OutgoingAmount:  pointer.ToString(moneyRequest.Amount.String()),
	}
	details, err := t.requestCreator.EvaluateTBURequest(&tbuForm, initiator)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))

		return
	}

	detail, ok := details[transactionConstants.PurposeTBUIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeTBUIncoming is not set"})
		return
	}

	destinationUser, err := t.userService.GetByUID(destinationAcc.UserId)
	if nil != err {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "Can not get a destination user details", OriginalError: err})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{
		Recipient: &recipient{
			PhoneNumber: destinationUser.PhoneNumber,
			FirstName:   destinationUser.FirstName,
			LastName:    destinationUser.LastName,
		},
		Details:              details,
		IncomingAmount:       detail.Amount.String(),
		IncomingCurrencyCode: detail.CurrencyCode,
	}))
}

func (t *MoneyRequestTbuHandler) CreateRequestUser(c *gin.Context) {
	initiator := t.contextService.MustGetCurrentUser(c)
	requestForm := &form.MoneyRequestTBUPreview{}

	if err := c.ShouldBind(requestForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	moneyRequest, typedErr := t.moneyRequestService.GetByTargetUID(*requestForm.MoneyRequestId, initiator.UID)
	if typedErr != nil {
		errorsPkg.AddShouldBindError(c, typedErr)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*requestForm.AccountIdFrom)
	if err != nil {
		t.logger.Info("tbuHandler unable to find account %d: %s", *requestForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	destinationAcc, err := t.accountRepository.FindByID(moneyRequest.RecipientAccountID)
	if err != nil {
		t.logger.Info("tbuHandler unable to find account %d: %s", moneyRequest.RecipientAccountID, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != initiator.UID || destinationAcc.UserId == initiator.UID {
		errcodes.AddError(c, errcodes.CodeForbidden)
		return
	}

	tbuForm := form.TBU{
		AccountIdFrom:   requestForm.AccountIdFrom,
		AccountNumberTo: &destinationAcc.Number,
		OutgoingAmount:  pointer.ToString(moneyRequest.Amount.String()),
		Description:     &moneyRequest.Description,
		IncomingAmount:  pointer.ToString(moneyRequest.Amount.String()),
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

	if !detail.Amount.Equal(moneyRequest.Amount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateTBURequest(&tbuForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	moneyRequest.Status = model.StatusApproved
	moneyRequest.IsNew = false
	moneyRequest.RequestID = req.Id
	if err := t.moneyRequestService.Update(moneyRequest, tx); err != nil {
		tx.Rollback()
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}
