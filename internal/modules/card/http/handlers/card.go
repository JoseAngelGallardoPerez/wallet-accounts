package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	cardSerializer "github.com/Confialink/wallet-accounts/internal/modules/card/serializer"
	cardService "github.com/Confialink/wallet-accounts/internal/modules/card/service"
)

var createInputFields = []string{"Number", "Status", "CardTypeId", "UserId", "ExpirationYear", "ExpirationMonth"}
var createOutputFields = []interface{}{"Id", "Number", "Status", "Balance", "CardTypeId",
	"UserId", "ExpirationYear", "ExpirationMonth", "CreatedAt"}
var showOutputFields = []interface{}{"Id", "Number", "Status", "Balance", "CardTypeId",
	"UserId", "ExpirationYear", "ExpirationMonth", "CreatedAt",
	map[string][]interface{}{
		"CardType": {"Id", "Name", "CurrencyCode", "IconId", map[string][]interface{}{
			"Category": {"Id", "Name"},
		}, map[string][]interface{}{
			"Format": {"Id", "Name", "Code"},
		}},
		"User": {"Id", "Username", "Email", "FirstName", "LastName"},
	}}
var updateInputFields = []string{"Status", "ExpirationYear", "ExpirationMonth"}
var updateOutputFields = []interface{}{"Id", "Number", "Balance", "Status", "CardTypeId",
	"UserId", "ExpirationYear", "ExpirationMonth", "CreatedAt"}

type CardHandler struct {
	contextService appHttpService.ContextInterface
	serializer     cardSerializer.CardSerializerInterface
	service        *cardService.CardService
	csvService     *cardService.Csv
	logger         log15.Logger
}

func NewCardHandler(
	contextService appHttpService.ContextInterface,
	serializer cardSerializer.CardSerializerInterface,
	service *cardService.CardService,
	csvService *cardService.Csv,
	logger log15.Logger,
) *CardHandler {
	return &CardHandler{
		contextService,
		serializer,
		service,
		csvService,
		logger.New("Handler", "CardHandler"),
	}
}

// CreateCardHandler handles creating new cards
func (h *CardHandler) CreateCardHandler(c *gin.Context) {
	data, _ := c.GetRawData()
	serializedIn, err := h.serializer.Deserialize(&data, createInputFields)
	if err != nil {
		errors.AddErrors(c, &errors.PrivateError{Message: err.Error()})
		return
	}

	created, err := h.service.Create(serializedIn, h.contextService.MustGetCurrentUser(c))

	if err != nil {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	serialized := h.serializer.Serialize(created, createOutputFields)
	c.JSON(http.StatusCreated, response.New().SetData(serialized))
}

// ShowCardHandler shows card by id in URI
func (h *CardHandler) ShowCardHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	id64, _ := strconv.ParseUint(id, 10, 32)
	showParams, errs := h.getShowParams(c.Request.URL.RawQuery)
	if len(errs) != 0 {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errs)
		return
	}

	if loaded, err := h.service.Get(uint32(id64), showParams); err == nil {
		serialized := h.serializer.Serialize(loaded, showParams.GetOutputFields())
		c.JSON(http.StatusOK, response.New().SetData(serialized))
	} else {
		privateError := errors.PrivateError{Message: "can't retrieve card"}
		privateError.AddLogPair("error", err)
		privateError.AddLogPair("card id", id)

		errors.AddErrors(c, &privateError)
	}
}

// UpdateCardHandler updates card by passed id in URI path
func (h *CardHandler) UpdateCardHandler(c *gin.Context) {
	data, _ := c.GetRawData()
	serializedFields, err := h.serializer.DeserializeFields(&data, updateInputFields)
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
		privateError := errors.PrivateError{Message: "can't update card"}
		privateError.AddLogPair("err", err)
		privateError.AddLogPair("request data", data)
		errors.AddErrors(c, &privateError)
	}
}

// ImportCsvHandler imports list of accounts from CSV
func (h *CardHandler) ImportCsvHandler(c *gin.Context) {
	logger := h.logger.New("action", "ImportCsvHandler")

	file, _, err := c.Request.FormFile("file")

	if nil != err {
		logger.Error("can't get file", "err", err)
		errcodes.AddError(c, errcodes.CodeFileInvalid)
		return
	}

	defer file.Close()

	if nil != err {
		privateError := errors.PrivateError{Message: err.Error()}
		errors.AddErrors(c, &privateError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		privateError := errors.PrivateError{Message: "can't write file to buffer"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	accounts, err := h.csvService.CsvToCards(buf)

	if nil != err {
		logger.Error("can't convert csv data to cards structures", "err", err)
		c.Status(http.StatusBadRequest)
		errcodes.AddError(c, errcodes.CodeFileInvalid)
		return
	}

	_, err = h.service.BulkCreate(accounts)

	if nil != err {
		errors.AddErrors(c, errcodes.ConvertToTyped(err))
		return
	}

	c.JSON(http.StatusOK, response.New().AddMessage("Cards successfully imported"))
}

func (h *CardHandler) getShowParams(query string) (*list_params.Includes, []error) {
	params := list_params.NewIncludes(query)
	params.AllowSelectFields(showOutputFields)
	params.Allow([]string{"cardType", "user", "cardType.category", "cardType.format"})
	params.AddCustomIncludes("user", h.service.UserIncludes)
	if ok, errs := params.Validate(); !ok {
		return nil, errs
	}
	return params, nil
}
