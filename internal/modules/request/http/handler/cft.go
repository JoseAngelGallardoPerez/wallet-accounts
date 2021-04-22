package handler

import (
	"errors"
	"net/http"

	errorsPkg "github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	cardRepository "github.com/Confialink/wallet-accounts/internal/modules/card/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
)

type CftHandler struct {
	contextService    service.ContextInterface
	accountRepository *accountRepository.AccountRepository
	cardRepository    cardRepository.CardRepositoryInterface
	requestCreator    *request.Creator
	logger            log15.Logger
	db                *gorm.DB
}

func NewCftHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	cardRepository cardRepository.CardRepositoryInterface,
	requestCreator *request.Creator,
	db *gorm.DB,
	logger log15.Logger,
) *CftHandler {
	return &CftHandler{
		contextService:    contextService,
		accountRepository: accountRepository,
		cardRepository:    cardRepository,
		requestCreator:    requestCreator,
		logger:            logger.New("Handler", "CftHandler"),
		db:                db,
	}
}

func (h *CftHandler) CreatePreviewAdmin(c *gin.Context) {
	logger := h.logger.New("action", "CreatePreviewAdmin")
	ownerId := c.Param("userId")
	cftForm := &form.CFTPreview{}

	if err := c.ShouldBind(cftForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := h.accountRepository.FindByID(*cftForm.AccountIdFrom)
	if err != nil {
		logger.Error("failed to retrieve source account", "error", err, "accountId", *cftForm.AccountIdFrom)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	card, err := h.cardRepository.Get(*cftForm.CardIdTo, nil)
	if err != nil {
		logger.Error("failed to get card", "error", err)
		errcodes.AddError(c, errcodes.CodeCardNotFound)
		return
	}

	if *card.UserId != ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidCardOwner)
	}

	user := h.contextService.MustGetCurrentUser(c)
	details, err := h.requestCreator.EvaluateCFTRequest(cftForm, user)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeCFTIncoming]
	if !ok {
		privateError := errorsPkg.PrivateError{Message: "transaction detail PurposeCFTIncoming is not set"}
		privateError.AddLogPair("error", err)
		errorsPkg.AddErrors(c, &privateError)

		logger.Crit("logic error", "error", err)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, IncomingAmount: detail.Amount.String()}))
}

func (h *CftHandler) CreatePreviewUser(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)

	logger := h.logger.New("action", "CreatePreviewUser")
	cftForm := &form.CFTPreview{}

	if err := c.ShouldBind(cftForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := h.accountRepository.FindByID(*cftForm.AccountIdFrom)
	if err != nil {
		logger.Error("failed to retrieve source account", "error", err, "accountId", *cftForm.AccountIdFrom)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != user.UID {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	card, err := h.cardRepository.Get(*cftForm.CardIdTo, nil)
	if err != nil {
		logger.Error("failed to get card", "error", err)
		errcodes.AddError(c, errcodes.CodeCardNotFound)
		return
	}

	if *card.UserId != user.UID {
		err := errcodes.CreatePublicError(errcodes.CodeInvalidCardOwner, "card must belong to the same user")
		logger.Error("unable to create request", "error", err)

		errorsPkg.AddErrors(c, err)
		return
	}

	details, err := h.requestCreator.EvaluateCFTRequest(cftForm, user)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeCFTIncoming]
	if !ok {
		err := errors.New("transaction detail PurposeCFTIncoming is not set")
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: err.Error()})
		logger.Crit("logic error", "error", err)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, IncomingAmount: detail.Amount.String()}))
}

func (h *CftHandler) CreateRequestUser(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)
	logger := h.logger.New("action", "CreateRequestUser")

	cftForm := &form.CFT{}
	if err := c.ShouldBind(cftForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := h.accountRepository.FindByID(*cftForm.AccountIdFrom)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != user.UID {
		err := errors.New("account does not belong to the given user")
		logger.Error("failed to create cft request", "error", err)
		errcodes.AddError(c, errcodes.CodeForbidden)
		return
	}

	card, err := h.cardRepository.Get(*cftForm.CardIdTo, nil)
	if err != nil {
		logger.Error("failed to get card", "error", err)
		errcodes.AddError(c, errcodes.CodeCardNotFound)
		return
	}

	if *card.UserId != user.UID {
		errcodes.AddError(c, errcodes.CodeInvalidCardOwner)
	}

	details, err := h.requestCreator.EvaluateCFTRequest(cftForm.ToCFTPreview(), user)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeCFTIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeCFTIncoming is not set"})
		return
	}

	formIncomingAmount, _ := decimal.NewFromString(*cftForm.IncomingAmount)
	if !detail.Amount.Equal(formIncomingAmount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := h.db.Begin()
	req, err := h.requestCreator.CreateCFTRequest(cftForm, user, tx)
	if err != nil {
		tx.Rollback()
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (h *CftHandler) CreateRequestAdmin(c *gin.Context) {
	logger := h.logger.New("action", "CreateRequestAdmin")

	ownerId := c.Param("userId")
	initiator := h.contextService.MustGetCurrentUser(c)
	cftForm := &form.CFT{}

	if err := c.ShouldBind(cftForm); err != nil {
		errorsPkg.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := h.accountRepository.FindByID(*cftForm.AccountIdFrom)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != ownerId {
		err := errors.New("account does not belong to the given user")
		logger.Error("failed to create owt request", "error", err)
		errcodes.AddError(c, errcodes.CodeForbidden)
		return
	}

	card, err := h.cardRepository.Get(*cftForm.CardIdTo, nil)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeCardNotFound)
		return
	}

	if *card.UserId != ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidCardOwner)
		return
	}

	user := h.contextService.MustGetCurrentUser(c)
	details, err := h.requestCreator.EvaluateCFTRequest(cftForm.ToCFTPreview(), user)
	if err != nil {
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeCFTIncoming]
	if !ok {
		errorsPkg.AddErrors(c, &errorsPkg.PrivateError{Message: "transaction detail PurposeCFTIncoming is not set"})
		return
	}

	formIncomingAmount, _ := decimal.NewFromString(*cftForm.IncomingAmount)
	if !detail.Amount.Equal(formIncomingAmount) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := h.db.Begin()
	req, err := h.requestCreator.CreateCFTRequest(cftForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errorsPkg.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}
