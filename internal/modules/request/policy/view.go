package policy

import (
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/policy"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	transactionRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"
)

func ProvideClientViewRequestPolicy(
	transactionRepository *transactionRepository.TransactionRepository,
	accountRepository *accountRepository.AccountRepository,
	logger log15.Logger,
) policy.Policy {
	return func(request interface{}, user *users.User) bool {
		req := request.(*model.Request)
		transactions, err := transactionRepository.GetByRequestId(*req.Id)
		if err != nil {
			logger.Error("failed to get transactions by request id", "error", err)
			return false
		}

		for _, transaction := range transactions {
			if transaction.AccountId != nil {
				account, err := accountRepository.FindByID(*transaction.AccountId)
				if err != nil {
					logger.Error("failed to get account by id", "error", err)
					return false
				}
				if account.UserId == user.UID {
					return true
				}
			}
		}
		return false
	}
}
