package handler

import (
	"net/http"

	transaction_view "github.com/Confialink/wallet-accounts/internal/modules/transaction/transaction-view"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-pkg-errors"

	log "github.com/inconshreveable/log15"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/response"
	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	transactionsRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
)

type HistoryHandler struct {
	contextService         service.ContextInterface
	logger                 log.Logger
	transactionsRepository *transactionsRepository.TransactionRepository
	txView                 transaction_view.View
	params                 *HandlerParams
}

func NewHistoryHandler(
	contextService service.ContextInterface,
	logger log.Logger,
	transactionsRepository *transactionsRepository.TransactionRepository,
	txView transaction_view.View,
	params *HandlerParams,
) *HistoryHandler {
	txSerializer := transaction_view.UserTxSerializer
	txView = txView.SetTransactionSerializer(txSerializer)
	return &HistoryHandler{
		contextService:         contextService,
		logger:                 logger.New("Handler", "transaction.HistoryHandler"),
		transactionsRepository: transactionsRepository,
		txView:                 txView,
		params:                 params,
	}
}

func (h *HistoryHandler) ShowUserHistory(c *gin.Context) {
	user := h.contextService.MustGetCurrentUser(c)
	logger := h.logger.New("action", "ShowUserHistory")

	listParams := h.params.forClientHistory(c.Request.URL.RawQuery)
	if ok, errorsList := listParams.Validate(); !ok {
		errcodes.AddErrorMeta(c, errcodes.CodeInvalidQueryParameters, errorsList)
		return
	}

	listParams.AddLeftJoin("accounts", "transactions.account_id = accounts.id")
	listParams.AddFilter("accounts.user_id", []string{user.UID})
	listParams.AddFilter("is_visible", []string{"1"})

	items, err := h.transactionsRepository.GetList(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "history handler. Unable to find transactions."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	count, err := h.transactionsRepository.GetListCount(listParams)
	if err != nil {
		privateError := errors.PrivateError{Message: "history handler. Unable to get count of transactions."}
		privateError.AddLogPair("error", err.Error())
		errors.AddErrors(c, &privateError)
		return
	}

	view, err := h.txView.View(items, user, "sender", "recipient", "fees", "fiat")
	if err != nil {
		logger.Error("failed to prepare transactions view", "error", err)
		c.Error(err)
		return
	}

	resp := response.NewWithListAndLinksAndPagination(view["transactions"], count, c.Request.URL.RequestURI(), listParams)
	c.JSON(http.StatusOK, resp)
}
