package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/payment-period/repository"
)

// AccountHandler
type PaymentPeriodHandler struct {
	repo           repository.PaymentPeriodRepositoryInterface
	contextService appHttpService.ContextInterface
	logger         log15.Logger
}

// NewPaymentPeriodHandler creates new payment method service
func NewPaymentPeriodHandler(
	repo repository.PaymentPeriodRepositoryInterface,
	contextService appHttpService.ContextInterface,
	logger log15.Logger,
) *PaymentPeriodHandler {
	return &PaymentPeriodHandler{
		repo:           repo,
		contextService: contextService,
		logger:         logger.New("Handler", "PaymentPeriodHandler"),
	}
}

// ListHandler returns the list of account types
func (h PaymentPeriodHandler) ListHandler(c *gin.Context) {
	items, err := h.repo.FindByParams(c.Request.URL.Query())

	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of payment periods"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
	r := response.NewWithList(items)
	c.JSON(http.StatusOK, r)
}
