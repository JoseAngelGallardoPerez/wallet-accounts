package handler

import (
	csvManager "github.com/Confialink/wallet-accounts/internal/modules/transaction/service/csv"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-utils/csv"
	"github.com/gin-gonic/gin"
	log "github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	transactionsRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	csvService "github.com/Confialink/wallet-accounts/internal/modules/transaction/service"
)

type CsvHandler struct {
	contextService         service.ContextInterface
	transactionsRepository *transactionsRepository.TransactionRepository
	csvService             *csvService.Csv
	params                 *HandlerParams
	logger                 log.Logger
	requestsMapper         *csvManager.RequestsMapper
}

func NewCsvHandler(
	contextService service.ContextInterface,
	transactionsRepository *transactionsRepository.TransactionRepository,
	csvService *csvService.Csv,
	params *HandlerParams,
	logger log.Logger,
	requestsMapper *csvManager.RequestsMapper,
) *CsvHandler {
	return &CsvHandler{
		contextService:         contextService,
		transactionsRepository: transactionsRepository,
		csvService:             csvService,
		params:                 params,
		logger:                 logger.New("Handler", "transaction.CsvHandler"),
		requestsMapper:         requestsMapper,
	}
}

func (h *CsvHandler) ExportToCsv(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)
	logger := h.logger.New("action", "ExportToCsv")

	listParams := h.params.forUserCsv(c.Request.URL.RawQuery)
	listParams.AddFilter("accounts.user_id", []string{user.UID})
	listParams.AddFilter("is_visible", []string{"1"})
	listParams.AddFilter("incomingByStatus", []string{transactionModel.StatusExecuted})
	listParams.AddLeftJoin("accounts", "transactions.account_id = accounts.id")
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	items, err := h.transactionsRepository.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "list handler. Unable to find transactions."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	file, err := h.csvService.FileWithTransactions(items)
	if err != nil {
		logger.Error("failed to prepare transactions file", "error", err)
		privateError := errors.PrivateError{Message: "Can not get csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	if err = csv.Send(file, c.Writer); err != nil {
		logger.Error("failed to send transactions file", "error", err)
		privateError := errors.PrivateError{Message: "Can not send csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
}

func (h *CsvHandler) ExportHistoryToCsv(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)
	logger := h.logger.New("action", "ExportHistoryToCsv")

	purpose := c.Request.URL.Query().Get("filter[purpose]")

	listParams := h.params.forClientHistoryCsv(c.Request.URL.RawQuery)
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	listParams.AddLeftJoin("accounts", "transactions.account_id = accounts.id")
	listParams.AddFilter("accounts.user_id", []string{user.UID})
	listParams.AddFilter("is_visible", []string{"1"})

	items, err := h.transactionsRepository.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "list handler. Unable to find transactions."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	requestsMap, err := h.requestsMapper.RequestsMap(items)
	if err != nil {
		logger.Error("failed to prepare requests map", "error", err)
		privateError := errors.PrivateError{Message: "Can not create requests map"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	file := csv.NewFile()
	doc := csvManager.NewProducer(purpose, file)

	err = doc.Construct(items, requestsMap)
	if err != nil {
		logger.Error("failed to prepare transactions file", "error", err)
		privateError := errors.PrivateError{Message: "Can not get csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	if err = csv.Send(file, c.Writer); err != nil {
		logger.Error("failed to send transactions file", "error", err)
		privateError := errors.PrivateError{Message: "Can not send csv file"}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}
}
