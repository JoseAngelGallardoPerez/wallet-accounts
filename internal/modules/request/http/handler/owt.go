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
)

type OwtHandler struct {
	contextService    service.ContextInterface
	accountRepository *accountRepository.AccountRepository
	requestCreator    *request.Creator
	logger            log15.Logger
	db                *gorm.DB
}

func NewOwtHandler(
	contextService service.ContextInterface,
	accountRepository *accountRepository.AccountRepository,
	requestCreator *request.Creator,
	db *gorm.DB,
	logger log15.Logger,

) *OwtHandler {
	return &OwtHandler{
		contextService:    contextService,
		accountRepository: accountRepository,
		requestCreator:    requestCreator,
		logger:            logger.New("Handler", "OwtHandler"),
		db:                db,
	}
}

func (t *OwtHandler) CreatePreviewAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreatePreviewAdmin")
	ownerId := c.Param("userId")
	owtForm := &form.OWTPreview{}

	if err := c.ShouldBind(owtForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*owtForm.AccountIdFrom)
	if err != nil {
		logger.Error("failed to retrieve source account", "error", err, "accountId", *owtForm.AccountIdFrom)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != ownerId {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateOWTRequest(owtForm, user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	totalOutgoingAmount := details.SumByAccountId(sourceAcc.ID)
	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, TotalOutgoingAmount: totalOutgoingAmount.String()}))
}

func (t *OwtHandler) CreatePreviewUser(c *gin.Context) {
	user := t.contextService.MustGetCurrentUser(c)
	logger := t.logger.New("action", "CreatePreviewAdmin")
	owtForm := &form.OWTPreview{}

	if err := c.ShouldBind(owtForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*owtForm.AccountIdFrom)
	if err != nil {
		logger.Error("failed to retrieve source account", "error", err, "accountId", *owtForm.AccountIdFrom)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != user.UID {
		errcodes.AddError(c, errcodes.CodeInvalidAccountOwner)
		return
	}

	details, err := t.requestCreator.EvaluateOWTRequest(owtForm, user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	totalOutgoingAmount := details.SumByAccountId(sourceAcc.ID)
	c.JSON(http.StatusOK, response.New().SetData(&preview{Details: details, TotalOutgoingAmount: totalOutgoingAmount.String()}))
}

func (t *OwtHandler) CreateRequestUser(c *gin.Context) {
	user := t.contextService.MustGetCurrentUser(c)

	logger := t.logger.New("action", "CreateRequestUser")

	owtForm := &form.OWT{}

	if err := c.ShouldBind(owtForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*owtForm.AccountIdFrom)
	if err != nil {
		logger.Error("failed to retrieve account", "error", err, "accountId", *owtForm.AccountIdFrom)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != user.UID {
		err := errcodes.CreatePublicError(errcodes.CodeInvalidAccountOwner, "account does not belong to the given user")
		logger.Error("failed to create owt request", "error", err)
		errors.AddErrors(c, err)
		return
	}

	details, err := t.requestCreator.EvaluateOWTRequest(owtForm.ToOWTPreview(), user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	totalOutgoingAmount := details.SumByAccountId(*owtForm.AccountIdFrom)
	confirmedOutgoingAmount, _ := decimal.NewFromString(*owtForm.ConfirmTotalOutgoingAmount)
	if !totalOutgoingAmount.Equal(confirmedOutgoingAmount) {
		logger.Info("owt request rejected", "cause", "confirmation amount does not match outgoing amount")
		err := errcodes.CreatePublicError(errcodes.CodeAmountsDoNatMatch, "confirmation amount does not match outgoing amount")
		errors.AddErrors(c, err)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateOWTRequest(owtForm, user, tx)
	if err != nil {
		tx.Rollback()
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}

func (t *OwtHandler) CreateRequestAdmin(c *gin.Context) {
	logger := t.logger.New("action", "CreateRequestAdmin")

	ownerId := c.Param("userId")

	initiator := t.contextService.MustGetCurrentUser(c)
	owtForm := &form.OWT{}

	if err := c.ShouldBind(owtForm); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	sourceAcc, err := t.accountRepository.FindByID(*owtForm.AccountIdFrom)
	if err != nil {
		logger.Error("failed to retrieve account", "error", err, "accountId", *owtForm.AccountIdFrom)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	if sourceAcc.UserId != ownerId {
		err := errcodes.CreatePublicError(errcodes.CodeInvalidAccountOwner, "account does not belong to the given user")
		logger.Error("failed to create owt request", "error", err)
		errors.AddErrors(c, err)
		return
	}

	user := t.contextService.MustGetCurrentUser(c)
	details, err := t.requestCreator.EvaluateOWTRequest(owtForm.ToOWTPreview(), user)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	totalOutgoingAmount := details.SumByAccountId(*owtForm.AccountIdFrom)
	confirmedOutgoingAmount, _ := decimal.NewFromString(*owtForm.ConfirmTotalOutgoingAmount)
	if !totalOutgoingAmount.Equal(confirmedOutgoingAmount) {
		logger.Info("owt request rejected", "cause", "confirmation amount does not match outgoing amount")
		err := errcodes.CreatePublicError(errcodes.CodeAmountsDoNatMatch, "confirmation amount does not match outgoing amount")
		errors.AddErrors(c, err)
		return
	}

	tx := t.db.Begin()
	req, err := t.requestCreator.CreateOWTRequest(owtForm, initiator, tx)
	if err != nil {
		tx.Rollback()
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.New().SetData(req))
}
