package system_logs

import (
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/recovery"

	requestConstants "github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	transactionsRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	usersService "github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/inconshreveable/log15"
)

type TransactionLogCreator struct {
	logsServiceWrap        *logsServiceWrap
	usersService           *usersService.UserService
	transactionsRepository *transactionsRepository.TransactionRepository
	logger                 log15.Logger
	transactionFinder      *TransactionFinder
	recoverer              func()
}

func NewTransactionLogCreator(usersService *usersService.UserService,
	transactionsRepository *transactionsRepository.TransactionRepository,
	logger log15.Logger,
	transactionFinder *TransactionFinder,
) *TransactionLogCreator {
	logger = logger.New("service", "SystemLogsService")
	return &TransactionLogCreator{
		logsServiceWrap:        newLogsServiceWrap(logger.New("serviceWrap", "logsServiceWrap")),
		usersService:           usersService,
		transactionsRepository: transactionsRepository,
		logger:                 logger,
		transactionFinder:      transactionFinder,
		recoverer:              recovery.DefaultRecoverer(),
	}
}

func (t *TransactionLogCreator) LogManualTransaction(request *model.Request) {
	user, err := t.usersService.GetByUID(*request.UserId)
	if err != nil {
		t.logger.Error("Failed to get user", "error", err)
		return
	}

	transactions, err := t.transactionsRepository.GetByRequestIdWithAccountAndType(*request.Id)
	if err != nil {
		t.logger.Error("Failed to get transactions", "error", err)
		return
	}
	var transaction *transactionModel.Transaction
	var operationType string

	switch *request.Subject {
	case requestConstants.SubjectDebitAccount:
		transaction = t.transactionFinder.getTransactionByPurpose(transactions, transactionConstants.PurposeDebitAccount.String())
		operationType = OperationTypeDebit
	case requestConstants.SubjectCreditAccount:
		transaction = t.transactionFinder.getTransactionByPurpose(transactions, transactionConstants.PurposeCreditAccount.String())
		operationType = OperationTypeCredit
	default:
		t.logger.Error("Can not log manual transaction. Unprocessable subject", "subject", *request.Subject)
	}

	if transaction == nil {
		t.logger.Error("Transaction is not found", "Request ID", *request.Id, "Operation type", operationType, "Transactions", transactions)
		return
	}

	t.createManualTransactionLog(transaction, user.UID, operationType, *request.BaseCurrencyCode)
}

func (t *TransactionLogCreator) LogRevenueManualTransaction(request *model.Request) {
	user, err := t.usersService.GetByUID(*request.UserId)
	if err != nil {
		t.logger.Error("Failed to get user", "error", err)
		return
	}

	transactions, err := t.transactionsRepository.GetByRequestIdWithAccounts(*request.Id)
	if err != nil {
		t.logger.Error("Failed to get transactions", "error", err)
		return
	}
	var transaction *transactionModel.Transaction
	var operationType string

	switch *request.Subject {
	case requestConstants.SubjectDebitRevenueAccount:
		transaction = t.transactionFinder.getTransactionByPurpose(transactions, transactionConstants.PurposeDebitRevenue.String())
		operationType = OperationTypeDebit
	default:
		t.logger.Error("Can not log manual transaction. Unprocessable subject", "subject", *request.Subject)
	}

	if transaction == nil {
		t.logger.Error("Transaction is not found", "Request ID", *request.Id, "Operation type", operationType, "Transactions", transactions)
		return
	}

	t.createRevenueManualTransactionLog(transaction, user.UID, operationType, *request.BaseCurrencyCode)
}

func (t *TransactionLogCreator) createManualTransactionLog(
	transaction *transactionModel.Transaction, userId, operationType string, currencyCode string,
) {
	defer t.recoverer()

	data, err := json.Marshal(transaction)
	if err != nil {
		t.logger.Error("Can't marshall json", err)
		return
	}

	t.logsServiceWrap.createLog(
		SubjectManualTransaction,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleMessage,
		data,
	)
}

func (t *TransactionLogCreator) createRevenueManualTransactionLog(
	transaction *transactionModel.Transaction, userId, operationType string, currencyCode string,
) {
	defer t.recoverer()

	data, err := json.Marshal(transaction)
	if err != nil {
		t.logger.Error("Can't marshall json", err)
		return
	}

	t.logsServiceWrap.createLog(
		SubjectRevenueDeduction,
		userId,
		time.Now().Format(time.RFC3339),
		DataTitleMessage,
		data,
	)
}
