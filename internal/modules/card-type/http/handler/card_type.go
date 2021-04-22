package handler

import (
	"net/http"
	"strconv"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/service"
)

var createInputFields = []string{"Name", "CurrencyCode", "IconId", "CardTypeCategoryId", "CardTypeFormatId"}
var createOutputFields = []interface{}{"Id", "Name", "CurrencyCode", "IconId", "CardTypeCategoryId", "CardTypeFormatId"}
var updateInputFields = []string{"Name", "CurrencyCode", "IconId", "CardTypeCategoryId", "CardTypeFormatId"}
var updateOutputFields = []interface{}{"Id", "Name", "CurrencyCode", "IconId", "CardTypeCategoryId", "CardTypeFormatId"}
var showOutputFields = []interface{}{"Id", "Name", "CurrencyCode", "IconId",
	map[string][]interface{}{
		"Category": {"Id", "Name"},
	}, map[string][]interface{}{
		"Format": {"Id", "Name", "Code"},
	}}

type CardTypeHandler struct {
	serializer     serializer.CardTypeSerializerInterface
	service        *service.CardTypeService
	params         *handlerParams
	logger         log15.Logger
	contextService appHttpService.ContextInterface
}

func NewCardTypeHandler(
	serializer serializer.CardTypeSerializerInterface,
	service *service.CardTypeService,
	logger log15.Logger,
	contextService appHttpService.ContextInterface,
) *CardTypeHandler {
	return &CardTypeHandler{
		serializer,
		service,
		&handlerParams{},
		logger.New("Handler", "CardTypeHandler"),
		contextService,
	}
}

// CreateHandler creates a cardType. Possible fields: "Name", "CurrencyCode", "IconId"
func (h *CardTypeHandler) CreateHandler(c *gin.Context) {
	data, _ := c.GetRawData()
	serializedIn, err := h.serializer.Deserialize(&data, createInputFields)
	if err != nil {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}

	if created, err := h.service.Create(serializedIn, h.contextService.MustGetCurrentUser(c)); err == nil {
		serialized := h.serializer.Serialize(created, createOutputFields)
		c.JSON(http.StatusCreated, response.New().SetData(serialized))
	} else {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
	}
}

// UpdateHandler updates cardType by id in URI. Possible fields: "Name", "CurrencyCode", "IconId"
func (h *CardTypeHandler) UpdateHandler(c *gin.Context) {
	data, _ := c.GetRawData()
	serializedFields, err := h.serializer.DeserializeFields(
		&data, updateInputFields)
	if err != nil {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}
	id := c.Params.ByName("id")
	id64, _ := strconv.ParseUint(id, 10, 32)

	if updated, err := h.service.UpdateFields(uint32(id64), serializedFields, h.contextService.MustGetCurrentUser(c)); err == nil {
		serialized := h.serializer.Serialize(updated, updateOutputFields)
		c.JSON(http.StatusOK, response.New().SetData(serialized))
	} else {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
	}
}

func (h *CardTypeHandler) ShowHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	id64, _ := strconv.ParseUint(id, 10, 32)
	includes := list_params.Includes{}
	includes.AddIncludes("format")
	includes.AddIncludes("category")
	if loaded, err := h.service.Get(uint32(id64), &includes); err == nil {
		serialized := h.serializer.Serialize(loaded, showOutputFields)
		c.JSON(http.StatusOK, response.New().SetData(serialized))
	} else {
		errcodes.AddError(c, errcodes.CodeCardTypeNotFound)
	}
}

// ListHandler returns list of records with pagination
func (h *CardTypeHandler) ListHandler(c *gin.Context) {
	listParams := h.params.list(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	items, err := h.service.GetList(listParams)
	count, countErr := h.service.GetCount(listParams)
	if err != nil || countErr != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of card types"}
		privateError.AddLogPair("error", err)
		errors.AddErrors(c, &privateError)
		return
	}

	serialized := h.serializer.SerializeList(items, listParams.GetOutputFields())
	resultResponse := response.NewWithListAndPageLinks(serialized, uint64(count),
		c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, resultResponse)
}

func (h *CardTypeHandler) DeleteHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	idU64, _ := strconv.ParseUint(id, 10, 32)

	if typedErr := h.service.Delete(uint32(idU64)); typedErr != nil {
		errors.AddErrors(c, typedErr)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
