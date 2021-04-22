package handler

import (
	"net/http"

	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

type CaHandler struct {
	contextService    service.ContextInterface
	accountRepository *accountRepository.AccountRepository
	requestCreator    *request.Creator
	db                *gorm.DB
	logger            log15.Logger
	systemLogger      *system_logs.SystemLogsService
}

func NewCaHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	requestCreator *request.Creator,
	db *gorm.DB,
	logger log15.Logger,
	systemLogger *system_logs.SystemLogsService,

) *CaHandler {
	return &CaHandler{
		contextService:    contextService,
		accountRepository: accountRepository,
		requestCreator:    requestCreator,
		db:                db,
		logger:            logger.New("Handler", "CaHandler"),
		systemLogger:      systemLogger,
	}
}

func (t *CaHandler) CreatePreviewAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreatePreviewAdmin")

	f := &form.CAPreview{}

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
	details, err := t.requestCreator.EvaluateCARequest(f, user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeCreditAccount]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeCreditAccount is not set"})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, IncomingAmount: detail.Amount.String()}))
}

func (t *CaHandler) CreateRequestAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreateRequestAdmin")

	initiator := t.contextService.MustGetCurrentUser(c)
	f := &form.CA{}

	if err := c.ShouldBind(f); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}
	_, err := t.accountRepository.FindByID(f.AccountId)
	if err != nil {
		logger.Error("CreateRequestAdmin unable to find account", f.AccountId, "err", err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateCARequest(f, initiator, tx)
	if err != nil {
		tx.Rollback()
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	tx.Commit()

	isAdmin, _ := t.requestCreator.GetIsAdminIsSystem(initiator)
	if isAdmin {
		t.systemLogger.LogManualTransactionAsync(req)
	}

	c.JSON(http.StatusOK, response.New().SetData(req))
}
