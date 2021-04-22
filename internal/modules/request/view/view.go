package view

import (
	accountsRepository "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	authService "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	transactionRepository "github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
	transactionView "github.com/Confialink/wallet-accounts/internal/modules/transaction/transaction-view"
	"github.com/Confialink/wallet-accounts/internal/modules/user"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/inconshreveable/log15"
)

type View interface {
	View(request *model.Request, currentUser *users.User) (map[string]interface{}, error)
}

type DefaultView struct {
	transactionRepository *transactionRepository.TransactionRepository
	transactionView       transactionView.View
	authService           authService.AuthServiceInterface
	userService           *service.UserService
	txSerializer          transactionView.TxSerializer
	accountsRepository    *accountsRepository.AccountRepository
	requestRepository     repository.RequestRepositoryInterface
	dataPresenter         request.DataPresenter
	logger                log15.Logger
}

func NewDefaultView(
	transactionRepository *transactionRepository.TransactionRepository,
	transactionView transactionView.View,
	authService authService.AuthServiceInterface,
	userService *service.UserService,
	txSerializer transactionView.TxSerializer,
	accountsRepository *accountsRepository.AccountRepository,
	requestRepository repository.RequestRepositoryInterface,
	dataPresenter request.DataPresenter,
	logger log15.Logger,
) View {
	return &DefaultView{
		transactionRepository: transactionRepository,
		transactionView:       transactionView,
		authService:           authService,
		userService:           userService,
		txSerializer:          txSerializer,
		accountsRepository:    accountsRepository,
		requestRepository:     requestRepository,
		dataPresenter:         dataPresenter,
		logger:                logger,
	}
}

func (b *DefaultView) View(request *model.Request, currentUser *users.User) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	requestResult := map[string]interface{}{
		"request": request,
	}
	result["request"] = requestResult

	if user.GetSystemUser().UID != *request.UserId {
		initiator, err := b.userService.GetByUID(*request.UserId)
		if err != nil {
			b.logger.Error("failed to fetch request initiator", "error", err.Error())
			return nil, err
		}

		requestResult["initiator"] = map[string]string{
			"UID":       initiator.GetUID(),
			"userName":  initiator.GetUsername(),
			"roleName":  initiator.GetRoleName(),
			"firstName": initiator.GetFirstName(),
			"lastName":  initiator.GetLastName(),
			"email":     initiator.GetEmail(),
		}
	}

	if destinationAccountId, ok := request.DestinationAccountId(); ok {
		destinationAccount, err := b.accountsRepository.FindByID(uint64(destinationAccountId))
		if err != nil {
			return nil, err
		}
		accUser, err := b.userService.GetByUID(destinationAccount.UserId)
		if err != nil {
			return nil, err
		}

		result["destinationAccount"] = map[string]interface{}{
			"number": destinationAccount.Number,
			"type": map[string]interface{}{
				"id":   destinationAccount.Type.ID,
				"name": destinationAccount.Type.Name,
			},
			"user": map[string]interface{}{
				"username":    accUser.Username,
				"firstName":   accUser.FirstName,
				"lastName":    accUser.LastName,
				"companyName": accUser.CompanyName,
				"email":       accUser.Email,
			},
		}
	}

	if sourceAccId, ok := request.SourceAccountId(); ok {
		sourceAccount, err := b.accountsRepository.FindByID(uint64(sourceAccId))
		if err != nil {
			return nil, err
		}
		accUser, err := b.userService.GetByUID(sourceAccount.UserId)
		if err != nil {
			return nil, err
		}

		result["sourceAccount"] = map[string]interface{}{
			"number": sourceAccount.Number,
			"type": map[string]interface{}{
				"id":   sourceAccount.Type.ID,
				"name": sourceAccount.Type.Name,
			},
			"user": map[string]interface{}{
				"username":    accUser.Username,
				"firstName":   accUser.FirstName,
				"lastName":    accUser.LastName,
				"companyName": accUser.CompanyName,
				"email":       accUser.Email,
			},
		}
	}

	data, err := b.dataPresenter.Present(request)
	if err != nil {
		b.logger.Warn("failed to present request data", "error", err, "requestId", *request.Id)
	}
	result["data"] = data

	transactions, err := b.transactionRepository.GetVisibleByRequestId(*request.Id)
	if err != nil {
		b.logger.Error("failed to get transactions by request id", "error", err, "requestId", *request.Id)
	}

	txViewPayload, err := b.transactionView.View(transactions, currentUser)
	if err != nil {
		b.logger.Error("failed to get render transaction view", "error", err)
		return nil, err
	}

	for key, value := range txViewPayload {
		result[key] = value
	}

	return result, nil
}
