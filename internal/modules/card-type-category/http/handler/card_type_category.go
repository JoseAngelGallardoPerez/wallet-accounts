package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/service"
)

type CardTypeCategoryHandler struct {
	service service.CardTypeCategoryServiceInterface
	logger  log15.Logger
}

func NewCardTypeCategoryHandler(
	service service.CardTypeCategoryServiceInterface,
	logger log15.Logger,
) *CardTypeCategoryHandler {
	return &CardTypeCategoryHandler{
		service: service,
		logger:  logger.New("Handler", "CardTypeCategoryHandler"),
	}
}

func (h *CardTypeCategoryHandler) ListHandler(c *gin.Context) {
	items, err := h.service.GetList()
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of card type categories"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusOK, response.NewWithList(items))
}
