package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	listParams "github.com/Confialink/wallet-pkg-list_params"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
)

// RevenueAccountHandler
type RevenueAccountHandler struct {
	repo           *repository.RevenueAccountRepository
	contextService appHttpService.ContextInterface
	service        *service.RevenueAccountService
	logger         log15.Logger
}

// NewRevenueAccountService creates new account service
func NewRevenueAccountHandler(
	repo *repository.RevenueAccountRepository,
	contextService appHttpService.ContextInterface,
	service *service.RevenueAccountService,
	logger log15.Logger,
) *RevenueAccountHandler {
	return &RevenueAccountHandler{
		repo:           repo,
		contextService: contextService,
		service:        service,
		logger:         logger.New("Handler", "RevenueAccountHandler"),
	}
}

// GetHandler returns account by id
func (h RevenueAccountHandler) GetHandler(c *gin.Context) {
	logger := h.logger.New("action", "GetHandler")

	id, typedErr := h.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}
	rev, err := h.repo.FindByID(id)
	if err != nil {
		logger.Error("can't retrieve revenue account", "err", err, "account id", id)
		errcodes.AddError(c, errcodes.CodeAccountNotFound)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(rev))
}

// ListHandler returns the list of revenue accounts
func (h RevenueAccountHandler) ListHandler(c *gin.Context) {
	params := h.getListParams(c.Request.URL.RawQuery)
	items, err := h.repo.GetList(params)
	if err != nil {
		privateError := errors.PrivateError{Message: "failed to retrieve count of revenue accounts"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
	count, countErr := h.repo.GetListCount()
	if countErr != nil {
		privateError := errors.PrivateError{Message: "failed to retrieve count of revenue accounts"}
		privateError.AddLogPair("error", countErr.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	r := response.NewWithListAndPageLinks(items, count, c.Request.URL.RequestURI(), params)
	c.JSON(http.StatusOK, r)
}

func (h *RevenueAccountHandler) getListParams(query string) *listParams.ListParams {
	params := listParams.NewListParamsFromQuery(query, model.Account{})
	params.AllowSelectFields(showListOutputFields)
	params.AllowPagination()
	return params
}
