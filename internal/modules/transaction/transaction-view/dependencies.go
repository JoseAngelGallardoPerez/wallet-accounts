package transaction_view

import (
	accountRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	requestRepository "github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	transactionRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/inconshreveable/log15"
)

var (
	requestsRepository     requestRepository.RequestRepositoryInterface
	accountsRepository     *accountRepository.AccountRepository
	usersService           *service.UserService
	requestDataPresenter   request.DataPresenter
	transactionsRepository *transactionRepository.TransactionRepository
	logger                 log15.Logger
)

func LoadDependencies(
	requestsRepositoryDep requestRepository.RequestRepositoryInterface,
	accountsRepositoryDep *accountRepository.AccountRepository,
	transactionsRepositoryDep *transactionRepository.TransactionRepository,
	requestDataPresenterDep request.DataPresenter,
	usersServiceDep *service.UserService,
	loggerDep log15.Logger,
) {
	requestsRepository = requestsRepositoryDep
	accountsRepository = accountsRepositoryDep
	transactionsRepository = transactionsRepositoryDep
	requestDataPresenter = requestDataPresenterDep
	usersService = usersServiceDep
	logger = loggerDep
}
