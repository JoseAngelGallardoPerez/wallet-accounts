package handlers

import (
	"net/http"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	cardSerializer "github.com/Confialink/wallet-accounts/internal/modules/card/serializer"
	cardService "github.com/Confialink/wallet-accounts/internal/modules/card/service"
	currencyService "github.com/Confialink/wallet-accounts/internal/modules/currency/service"
)

type CardListHandler struct {
	contextService  appHttpService.ContextInterface
	serializer      cardSerializer.CardSerializerInterface
	service         *cardService.CardService
	currencyService currencyService.CurrenciesServiceInterface
	params          *HandlerParams
	logger          log15.Logger
}

func NewCardListHandler(
	contextService appHttpService.ContextInterface,
	serializer cardSerializer.CardSerializerInterface,
	service *cardService.CardService,
	currencyService currencyService.CurrenciesServiceInterface,
	logger log15.Logger,
) *CardListHandler {
	return &CardListHandler{
		contextService,
		serializer,
		service,
		currencyService,
		NewHandlerParams(service),
		logger.New("Handler", "CardListHandler"),
	}
}

func (h *CardListHandler) IndexOwnCardsHandler(c *gin.Context) {
	currentUser := h.contextService.MustGetCurrentUser(c)

	listParams := h.params.getOwnListParams(c.Request.URL.RawQuery)
	if ok, errors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errors)
		return
	}

	listParams.AddFilter("userId", []string{currentUser.UID})
	h.processListParams(listParams, c)
}

func (h *CardListHandler) IndexCardsHandler(c *gin.Context) {
	listParams := h.params.getListParams(c.Request.URL.RawQuery)
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	h.processListParams(listParams, c)
}

func (h *CardListHandler) processListParams(listParams *list_params.ListParams, c *gin.Context) {
	loaded, err := h.service.GetList(listParams)
	count, countErr := h.service.GetListCount(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of cards"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
	if countErr != nil {
		privateError := errors.PrivateError{Message: "can't retrieve count of cards"}
		privateError.AddLogPair("error", countErr.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	serialized := make([]map[string]interface{}, len(loaded))

	for i, v := range loaded {
		serialized[i] = h.serializer.Serialize(v, listParams.GetOutputFields())
	}
	r := response.NewWithListAndPageLinks(serialized, count, c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, r)
}
