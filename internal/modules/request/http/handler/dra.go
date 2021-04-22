package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/form"
	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
)

type DraHandler struct {
	contextService    service.ContextInterface
	revenueRepository *accountRepository.RevenueAccountRepository
	requestCreator    *request.Creator
	db                *gorm.DB
	logger            log15.Logger
	systemLogger      *system_logs.SystemLogsService
}

func NewDraHandler(
	contextService service.ContextInterface,
	revenueRepository *accountRepository.RevenueAccountRepository,
	requestCreator *request.Creator,
	db *gorm.DB,
	logger log15.Logger,
	systemLogger *system_logs.SystemLogsService,
) *DraHandler {
	return &DraHandler{
		contextService:    contextService,
		revenueRepository: revenueRepository,
		requestCreator:    requestCreator,
		db:                db,
		logger:            logger.New("Handler", "DraHandler"),
		systemLogger:      systemLogger,
	}
}

func (t *DraHandler) CreatePreviewAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreatePreviewAdmin")

	f := &form.DRAPreview{}

	if err := c.ShouldBind(f); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	_, err := t.revenueRepository.FindByID(f.RevenueAccountId)
	if err != nil {
		logger.Error("CreatePreviewAdmin unable to find revenue account", f.RevenueAccountId, "err", err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateDRARequest(f, user)

	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	detail, ok := details[transactionConstants.PurposeDebitRevenue]
	if !ok {
		errors.AddErrors(c, &errors.PrivateError{Message: "transaction detail PurposeDebitAccount is not set"})
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, IncomingAmount: detail.Amount.String()}))
}

func (t *DraHandler) CreateRequestAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreateRequestAdmin")

	initiator := t.contextService.MustGetCurrentUser(c)
	f := &form.DRA{}

	if err := c.ShouldBind(f); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	_, err := t.revenueRepository.FindByID(f.RevenueAccountId)
	if err != nil {
		logger.Error("deduct revenue account handler unable to find revenue account", f.RevenueAccountId, "err", err.Error())
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateDRARequest(f, initiator, tx)
	if err != nil {
		tx.Rollback()
		logger.Error("credit account handler failed to create deduct revenue request", "err ", err)
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	isAdmin, _ := t.requestCreator.GetIsAdminIsSystem(initiator)
	if err == nil && isAdmin {
		t.systemLogger.LogRevenueManualTransactionAsync(req)
	}

	c.JSON(http.StatusOK, response.New().SetData(req))
}
