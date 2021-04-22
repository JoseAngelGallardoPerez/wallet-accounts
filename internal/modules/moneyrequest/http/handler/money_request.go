package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/model"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/service"
)

type MoneyRequest struct {
	moneyRequestService *service.MoneyRequest
	contextService      appHttpService.ContextInterface
	params              *Params
	logger              log15.Logger
}

func NewMoneyRequest(
	moneyRequestService *service.MoneyRequest,
	contextService appHttpService.ContextInterface,
	params *Params,
	logger log15.Logger,
) *MoneyRequest {
	return &MoneyRequest{
		moneyRequestService: moneyRequestService,
		contextService:      contextService,
		params:              params,
		logger:              logger.New("Handler", "MoneyRequestHandler"),
	}
}

type Request struct {
	TargetUID          string `json:"targetUID" binding:"required"`
	RecipientAccountID uint64 `json:"recipientAccountId" binding:"required"`
	Amount             string `json:"amount" binding:"required,decimalGT=0"`
	Description        string `json:"description" binding:"omitempty,max=255"`
}

func (h *MoneyRequest) Create(c *gin.Context) {
	data := &Request{}

	if err := c.ShouldBindJSON(&data); err != nil {
		errors.AddShouldBindError(c, err)
		return
	}

	moneyRequest := model.MoneyRequest{
		TargetUserID:       data.TargetUID,
		RecipientAccountID: data.RecipientAccountID,
		Amount:             decimal.RequireFromString(data.Amount),
		Description:        data.Description,
		IsNew:              true,
		Status:             model.StatusPending,
	}
	user := h.contextService.GetCurrentUser(c)
	createdMoneyRequest, typedErr := h.moneyRequestService.Create(&moneyRequest, user)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	c.JSON(http.StatusCreated, response.New().SetData(createdMoneyRequest))
}

// Show returns a money request for a user who should pay it
func (h *MoneyRequest) Show(c *gin.Context) {
	id, typedErr := h.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}
	currentUser := h.contextService.GetCurrentUser(c)

	moneyRequest, typedErr := h.moneyRequestService.GetByTargetUID(id, currentUser.UID)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(moneyRequest))
}

// MarkOld changes "IsNew" field of the request
func (h *MoneyRequest) MarkOld(c *gin.Context) {
	id, typedErr := h.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}
	currentUser := h.contextService.GetCurrentUser(c)

	moneyRequest, typedErr := h.moneyRequestService.GetByTargetUID(id, currentUser.UID)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	moneyRequest.IsNew = false
	err := h.moneyRequestService.Update(moneyRequest, nil)
	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(moneyRequest))
}

// Incoming returns a list of incoming money requests
func (h *MoneyRequest) Incoming(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)

	listParams := h.params.forIncoming(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	listParams.AddFilter("targetUserId", []string{currentUser.UID})

	records, err := h.moneyRequestService.GetList(listParams)
	if err != nil {
		privateErr := errors.PrivateError{Message: "can't retrieve list of money requests"}
		privateErr.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	totalCount, countErr := h.moneyRequestService.GetCount(listParams)
	if countErr != nil {
		privateErr := errors.PrivateError{Message: "can't count money requests"}
		privateErr.AddLogPair("error", countErr.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	r := response.NewWithListAndPagination(records, uint64(totalCount), listParams)
	c.JSON(http.StatusOK, r)
}

// Outgoing returns a list of outgoing money requests
func (h *MoneyRequest) Outgoing(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)

	listParams := h.params.forOutgoing(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	listParams.AddFilter("initiatorUserId", []string{currentUser.UID})

	records, err := h.moneyRequestService.GetList(listParams)
	if err != nil {
		privateErr := errors.PrivateError{Message: "can't retrieve list of money requests"}
		privateErr.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	totalCount, countErr := h.moneyRequestService.GetCount(listParams)
	if countErr != nil {
		privateErr := errors.PrivateError{Message: "can't count money requests"}
		privateErr.AddLogPair("error", countErr.Error())
		errors.AddErrors(c, &privateErr)
		return
	}

	r := response.NewWithListAndPagination(records, uint64(totalCount), listParams)
	c.JSON(http.StatusOK, r)
}
