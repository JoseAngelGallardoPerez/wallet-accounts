package handler

import (
	"log"
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
)

type TbaHandler struct {
	contextService    service.ContextInterface
	accountRepository *accountRepository.AccountRepository
	requestCreator    *request.Creator
	db                *gorm.DB
}

func NewTbaHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	requestCreator *request.Creator,
	db *gorm.DB,

) *TbaHandler {
	return &TbaHandler{
		contextService:    contextService,
		accountRepository: accountRepository,
		requestCreator:    requestCreator,
		db:                db,
	}
}

func (t *TbaHandler) CreatePreviewAdmin(c *gin.Context) {
	ownerId := c.Param("userId")

	tbaForm := &form.TBAPreview{}

	if err := c.ShouldBind(tbaForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}
	destinationAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdTo)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != ownerId || destinationAcc.UserId != ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateTBARequest(tbaForm, user)

	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeTBAIncoming]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeTBAIncoming is not TAN is required and cannot be empty set"})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, IncomingAmount: detail.Amount.String()}))
}

func (t *TbaHandler) CreatePreviewUser(c *gin.Context) {
	initiator := t.contextService.MustGetCurrentUser(c)
	tbaForm := &form.TBAPreview{}

	if err := c.ShouldBind(tbaForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}
	destinationAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdTo)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != initiator.UID || destinationAcc.UserId != initiator.UID {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	details, err := t.requestCreator.EvaluateTBARequest(tbaForm, initiator)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	incomingDetail, ok := details[transactionConstants.PurposeTBAIncoming]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeTBAIncoming is not set"})
		return
	}
	_, ok = details[transactionConstants.PurposeTBAOutgoing]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeTBAIncoming is not set"})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{
		Details: details, IncomingAmount: incomingDetail.Amount.String(),
		TotalOutgoingAmount: details.SumByAccountId(*tbaForm.AccountIdFrom).String(),
	}))
}

func (t *TbaHandler) CreateRequestUser(c *gin.Context) {
	initiator := t.contextService.MustGetCurrentUser(c)
	tbaForm := &form.TBA{}

	if err := c.ShouldBind(tbaForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}
	destinationAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdTo)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != initiator.UID || destinationAcc.UserId != initiator.UID {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	formIncomingAmount, _ := decimal.NewFromString(*tbaForm.IncomingAmount)

	details, err := t.requestCreator.EvaluateTBARequest(tbaForm.ToTBAPreview(), initiator)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeTBAIncoming]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeTBAIncoming is not set"})
		return
	}

	if !detail.Amount.Equal(formIncomingAmount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateTBARequest(tbaForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (t *TbaHandler) CreateRequestAdmin(c *gin.Context) {
	ownerId := c.Param("userId")

	initiator := t.contextService.MustGetCurrentUser(c)
	tbaForm := &form.TBA{}

	if err := c.ShouldBind(tbaForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdFrom)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}
	destinationAcc, err := t.accountRepository.FindByID(*tbaForm.AccountIdTo)
	if err != nil {
		log.Printf("tbaHandler unable to find account %d: %s", *tbaForm.AccountIdFrom, err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != ownerId || destinationAcc.UserId != ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	details, err := t.requestCreator.EvaluateTBARequest(tbaForm.ToTBAPreview(), initiator)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeTBAIncoming]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeTBAIncoming is not set"})
		return
	}

	formIncomingAmount, _ := decimal.NewFromString(*tbaForm.IncomingAmount)
	if !detail.Amount.Equal(formIncomingAmount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateTBARequest(tbaForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}
