package handler

import (
	"fmt"
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	scheduled_transaction "github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction"
	"github.com/Confialink/wallet-accounts/internal/modules/scheduled-transaction/service"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-model_serializer"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/gin-gonic/gin"
	"github.com/inconshreveable/log15"
)

type TransactionsHandler struct {
	repo           *scheduled_transaction.Repository
	contextService appHttpService.ContextInterface
	logger         log15.Logger
}

func NewTransactionsHandler(
	repo *scheduled_transaction.Repository,
	contextService appHttpService.ContextInterface,
	logger log15.Logger,
) *TransactionsHandler {
	return &TransactionsHandler{repo: repo,
		contextService: contextService,
		logger:         logger.New("Handler", "TransactionHandler"),
	}
}

func (h *TransactionsHandler) ListHandler(c *gin.Context) {
	listParams := getListParams(c.Request.URL.RawQuery)
	if ok, paramsErrors := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	items, err := h.repo.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "can't retrieve list of of IWT bank details"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
	count, countErr := h.repo.GetListCount(listParams)
	if countErr != nil {
		privateError := errors.PrivateError{Message: "can't retrieve count of IWT bank details"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	serialized := make([]interface{}, len(items))
	fields := listParams.GetOutputFields()
	for i, v := range items {
		fmt.Println(v)
		fmt.Println(fields)
		serialized[i] = model_serializer.Serialize(v, fields)
	}
	r := response.NewWithListAndPageLinks(serialized, count, c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, r)
}

// TransactionsHandler returns scheduled transaction by id
func (h *TransactionsHandler) GetById(c *gin.Context) {
	logger := h.logger.New("action", "GetByAccountIdHandler")

	id, terr := h.contextService.GetIdParam(c)
	if terr != nil {
		errcodes.AddError(c, errcodes.CodeInvalidQueryParameters)
		return
	}

	transaction, err := h.repo.FindByID(id)
	if err != nil {
		errcodes.AddError(c, errcodes.CodeTransactionNotFound)
		return
	}

	if err != nil {
		logger.Error("can't retrieve scheduled transaction", "err", err, "scheduled transaction id", id)
		errcodes.AddError(c, errcodes.CodeTransactionNotFound)
		return
	}

	c.JSON(http.StatusOK, response.New().SetData(transaction))
}

// ExportToCsv handler for downloading transactions csv
func (h *TransactionsHandler) ExportToCsv(c *gin.Context) {
	logger := h.logger.New("action", "ExportToCsv")
	params := getCsvParams(c.Request.URL.RawQuery)
	if ok, paramsErrors := params.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, paramsErrors)
		return
	}

	transactions, err := h.repo.GetList(params)
	if err != nil {
		logger.Error("can't retrieve scheduled transactions", "err", err)
		errcodes.AddError(c, errcodes.CodeTransactionNotFound)
		return
	}

	file, err := service.GetCsv(transactions)
	if err != nil {
		logger.Error("can't get csv file", "err", err)
		errcodes.AddError(c, errcodes.CodeTransactionNotFound)
		return
	}

	if err = csv.Send(file, c.Writer); err != nil {
		privateError := errors.PrivateError{Message: "Can not send csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
	}
}
