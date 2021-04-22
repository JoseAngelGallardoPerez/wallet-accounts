package handler

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/service"
)

type CardTypeFormatHandler struct {
	service service.CardTypeFormatServiceInterface
	logger  log15.Logger
}

func NewCardTypeFormatHandler(
	service service.CardTypeFormatServiceInterface,
	logger log15.Logger,
) *CardTypeFormatHandler {
	return &CardTypeFormatHandler{
		service: service,
		logger:  logger.New("Handler", "CardTypeFormatHandler"),
	}
}

func (h *CardTypeFormatHandler) ListHandler(c *gin.Context) {
	items, err := h.service.GetList()
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of card type formats"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	c.JSON(http.StatusOK, response.NewWithList(items))
}
