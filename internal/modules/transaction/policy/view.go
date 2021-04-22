package policy

import (
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/policy"
	transactionConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	transactionsRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/utils"
	"github.com/Confialink/wallet-pkg-utils/value"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"
)

var neverShownToClientPurposes = map[transactionConstants.Purpose]uint8{
	transactionConstants.PurposeCreditRevenue: 0,
	transactionConstants.PurposeDebitRevenue:  0,
}

func ProvideClientViewTransaction(
	accountRepository *accountRepository.AccountRepository,
	transactionsRepository *transactionsRepository.TransactionRepository,
	logger log15.Logger,
) policy.Policy {
	return func(transaction interface{}, user *users.User) bool {
		tx := transaction.(*model.Transaction)
		if _, neverShow := neverShownToClientPurposes[transactionConstants.Purpose(*tx.Purpose)]; neverShow {
			return false
		}
		if !value.FromBool(tx.IsVisible) {
			return false
		}
		return isBelongToOwnAccount(tx, user, accountRepository, logger) ||
			isRequestMainTransaction(tx, user, transactionsRepository, logger)
	}
}

// isBelongToOwnAccount checks if transaction belongs to account of passed user
func isBelongToOwnAccount(transaction *model.Transaction, user *users.User,
	accountRepository *accountRepository.AccountRepository, logger log15.Logger,
) bool {
	if transaction.AccountId == nil {
		return false
	}

	account, err := accountRepository.FindByID(*transaction.AccountId)
	if err != nil {
		logger.Error("failed to get account by id", "error", err)
		return false
	}

	return account.UserId == user.UID
}

// isRequestMainTransaction checks if transaction belongs to request where user has interest and it's not a revenue
func isRequestMainTransaction(transaction *model.Transaction, user *users.User,
	transactionsRepository *transactionsRepository.TransactionRepository,
	logger log15.Logger,
) bool {
	transactions, err := transactionsRepository.GetByRequestIdWithAccounts(*transaction.RequestId)
	if err != nil {
		logger.Error("failed to get transactions by request id", "error", err)
		return false
	}

	// If we have at least one transaction related to user in request
	// Then we check current transaction
	for _, tr := range transactions {
		if tr.Account != nil && tr.Account.UserId == user.UID {
			return utils.IsMainTransactionPurpose(transactionConstants.Purpose(*transaction.Purpose))
		}
	}

	return false
}
