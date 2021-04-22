package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/country/service"
)

// CountryHandler
type CountryHandler struct {
	contextService appHttpService.ContextInterface
	service        *service.CountryService
	logger         log15.Logger
}

// NewCountryService creates new account service
func NewCountryHandler(
	contextService appHttpService.ContextInterface,
	service *service.CountryService,
	logger log15.Logger,
) *CountryHandler {
	return &CountryHandler{
		contextService: contextService,
		service:        service,
		logger:         logger.New("Handler", "CountryHandler"),
	}
}

// ListAllHandler returns the list of all countries
func (h CountryHandler) ListAllHandler(c *gin.Context) {
	countries, err := h.service.FindAll()
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of countries"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	r := response.New().SetData(countries)
	c.JSON(http.StatusOK, r)
}

// ListAllHandler returns the list of all countries
func (h CountryHandler) GetHandler(c *gin.Context) {
	id, typedErr := h.contextService.GetIdParam(c)
	if typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}
	countries, err := h.service.FindById(pointer.ToUint(uint(id)))

	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve country"}
		privateError.AddLogPair("error", err.Error())
		privateError.AddLogPair("country id", id)
		errors.AddErrors(c, &privateError)
		return
	}

	r := response.New().SetData(countries)
	c.JSON(http.StatusOK, r)
}
