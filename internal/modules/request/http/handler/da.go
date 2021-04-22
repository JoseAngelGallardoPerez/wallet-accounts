package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
)

type DaHandler struct {
	contextService    service.ContextInterface
	accountRepository *accountRepository.AccountRepository
	requestCreator    *request.Creator
	logger            log15.Logger
	db                *gorm.DB
	systemLogger      *system_logs.SystemLogsService
}

func NewDaHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	requestCreator *request.Creator,
	db *gorm.DB,
	logger log15.Logger,
	systemLogger *system_logs.SystemLogsService,
) *DaHandler {
	return &DaHandler{
		contextService:    contextService,
		accountRepository: accountRepository,
		requestCreator:    requestCreator,
		logger:            logger.New("Handler", "DaHandler"),
		db:                db,
		systemLogger:      systemLogger,
	}
}

func (t *DaHandler) CreatePreviewAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreatePreviewAdmin")

	f := &form.DAPreview{}

	if err := c.ShouldBind(f); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	_, err := t.accountRepository.FindByID(f.AccountId)
	if err != nil {
		logger.Error("CreatePreviewAdmin unable to find account", f.AccountId, "err", err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateDARequest(f, user)

	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeDebitAccount]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeDebitAccount is not set"})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, IncomingAmount: detail.Amount.String()}))
}

func (t *DaHandler) CreateRequestAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreateRequestAdmin")

	initiator := t.contextService.MustGetCurrentUser(c)
	f := &form.DA{}

	if err := c.ShouldBind(f); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	_, err := t.accountRepository.FindByID(f.AccountId)
	if err != nil {
		logger.Error("credit account handler unable to find account", f.AccountId, "err", err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateDARequest(f.ToDAPreview(), user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeDebitAccount]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeDebitAccount is not set"})
		return
	}

	amount, _ := decimal.NewFromString(f.Amount)
	if !detail.Amount.Equal(amount.Neg()) {
		errcodes.AddError(c, errcodes.CodeRatesDoNotMatch)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateDARequest(f, initiator, tx)
	if err != nil {
		tx.Rollback()
		logger.Error("credit account handler failed to create credit request", "err ", err)
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	isAdmin, _ := t.requestCreator.GetIsAdminIsSystem(initiator)
	if err == nil && isAdmin {
		t.systemLogger.LogManualTransactionAsync(req)
	}

	c.JSON(http.StatusOK, response.New().SetData(req))
}
