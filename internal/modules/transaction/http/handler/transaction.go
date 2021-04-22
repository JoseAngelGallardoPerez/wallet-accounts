package handler

import (
	"net/http"

	transaction_view "github.com/Confialink/wallet-accounts/internal/modules/transaction/transaction-view"
	"github.com/Confialink/wallet-pkg-list_params"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
	log "github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	transactionsRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
)

type TransactionHandler struct {
	contextService         service.ContextInterface
	logger                 log.Logger
	transactionsRepository *transactionsRepository.TransactionRepository
	txView                 transaction_view.View
	txSerializer           transaction_view.TxSerializer
	params                 *HandlerParams
}

func NewTransactionHandler(
	contextService service.ContextInterface,
	logger log.Logger,
	transactionsRepository *transactionsRepository.TransactionRepository,
	txView transaction_view.View,
	params *HandlerParams,
) *TransactionHandler {
	txSerializer := transaction_view.UserTxSerializer
	txView = txView.SetTransactionSerializer(txSerializer)
	return &TransactionHandler{
		contextService:         contextService,
		logger:                 logger.New("Handler", "transaction.TransactionHandler"),
		transactionsRepository: transactionsRepository,
		txView:                 txView,
		txSerializer:           txSerializer,
		params:                 params,
	}
}

func (h *TransactionHandler) GetOne(c *gin.Context) {
	logger := h.logger.New("action", "GetOne")

	transaction := h.contextService.GetRequestedTransaction(c)
	if transaction == nil {
		logger.Error("unable to find transaction")
		return
	}

	includes := list_params.NewIncludes(c.Request.URL.RawQuery)
	includes.Allow([]string{"sender", "recipient", "requestData"})

	serializedTx, _ := h.txSerializer(transaction, includes.GetPreloads()...)
	resp := response.New().SetData(serializedTx)

	c.JSON(http.StatusOK, resp)
}

func (h *TransactionHandler) ShowUserList(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)
	logger := h.logger.New("action", "ShowUserList")

	listParams := h.params.forUser(c.Request.URL.RawQuery)
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

	count, err := h.transactionsRepository.GetListCount(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "list handler. Unable to get count of transactions."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	view, err := h.txView.View(items, user)
	if err != nil {
		logger.Error("failed to prepare transactions view", "error", err)
		_ = c.Error(err)
		return
	}

	resp := response.NewWithListAndPageLinks(view["transactions"], count, c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, resp)
}
