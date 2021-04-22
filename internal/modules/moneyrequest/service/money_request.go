package service

import (
	"fmt"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-users/rpc/proto/users"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountTypeModel "github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	accountService "github.com/Confialink/wallet-accounts/internal/modules/account/service"
	"github.com/Confialink/wallet-accounts/internal/modules/calculation"
	"github.com/Confialink/wallet-accounts/internal/modules/money"
	moneyRequestEvent "github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/event"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/model"
	"github.com/Confialink/wallet-accounts/internal/modules/moneyrequest/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/notifications"
	"github.com/Confialink/wallet-accounts/internal/modules/user/service"
)

type MoneyRequest struct {
	repo                *repository.MoneyRequest
	usersService        *service.UserService
	accounts            *accountService.AccountService
	rounding            *calculation.Rounding
	notificationService *notifications.Service
}

func NewMoneyRequest(repository *repository.MoneyRequest, usersService *service.UserService,
	accounts *accountService.AccountService, rounding *calculation.Rounding, notificationService *notifications.Service) *MoneyRequest {
	return &MoneyRequest{repository, usersService, accounts, rounding, notificationService}
}

func (s *MoneyRequest) Create(moneyRequest *model.MoneyRequest, currentUser *users.User) (*model.MoneyRequest, errors.TypedError) {
	account, err := s.accounts.FindByID(moneyRequest.RecipientAccountID)
	if err != nil {
		return nil, &errors.PrivateError{OriginalError: err}
	}
	moneyRequest.InitiatorUserID = currentUser.UID
	moneyRequest.CurrencyCode = account.Type.CurrencyCode
	if typedErr := s.validateMoneyRequest(moneyRequest, account, currentUser); typedErr != nil {
		return nil, typedErr
	}

	if err = s.repo.Create(moneyRequest); err != nil {
		return nil, &errors.PrivateError{OriginalError: err}
	}

	eventContext := &moneyRequestEvent.Context{
		MoneyRequestId:  moneyRequest.ID,
		RecipientUID:    moneyRequest.TargetUserID,
		SenderFirstName: currentUser.FirstName,
		SenderLastName:  currentUser.LastName,
		Amount:          moneyRequest.Amount,
		Currency:        account.Type.CurrencyCode,
	}

	_ = s.notificationService.TriggerNewMoneyRequest(eventContext)

	return moneyRequest, nil
}

func (s *MoneyRequest) GetByTargetUID(id uint64, targetUID string) (*model.MoneyRequest, errors.TypedError) {
	request, err := s.repo.GetByTargetUID(id, targetUID)
	if err != nil {
		return nil, &errors.PrivateError{OriginalError: err}
	}

	account, err := s.accounts.FindByID(request.RecipientAccountID)
	if err != nil {
		return nil, &errors.PrivateError{OriginalError: err}
	}

	recipient, err := s.usersService.GetByUID(account.UserId)
	if err != nil {
		return nil, &errors.PrivateError{OriginalError: err}
	}

	request.Recipient = &model.User{
		FirstName:   recipient.FirstName,
		LastName:    recipient.LastName,
		PhoneNumber: recipient.PhoneNumber,
	}
	return request, nil
}

func (s *MoneyRequest) Update(moneyRequest *model.MoneyRequest, tx *gorm.DB) error {
	repo := s.repo
	if tx != nil {
		repo = s.repo.WrapContext(tx)
	}

	return repo.Update(moneyRequest)
}

func (s *MoneyRequest) GetList(listParams *list_params.ListParams) (
	[]*model.MoneyRequest, error) {
	return s.repo.GetList(listParams)
}

func (s *MoneyRequest) GetCount(listParams *list_params.ListParams) (
	int64, error) {
	return s.repo.GetCount(listParams)
}

func (s *MoneyRequest) validateMoneyRequest(moneyRequest *model.MoneyRequest, account *accountModel.Account, currentUser *users.User) errors.TypedError {
	_, err := s.usersService.GetByUID(moneyRequest.TargetUserID)
	if err != nil {
		return &errors.PrivateError{OriginalError: err}
	}

	if account.UserId != currentUser.UID {
		return errcodes.CreatePublicError(errcodes.CodeInvalidAccountOwner)
	}

	if typedErr := s.validateTargetUserCurrency(moneyRequest.TargetUserID, account.Type.CurrencyCode); typedErr != nil {
		return typedErr
	}

	return s.rounding.ValidatePrecision(money.Amount{
		Value:        moneyRequest.Amount,
		CurrencyCode: account.Type.CurrencyCode,
	})
}

func (s *MoneyRequest) validateTargetUserCurrency(targetUID, currency string) errors.TypedError {
	accountTypeTableName := (&accountTypeModel.AccountType{}).TableName()
	params := list_params.NewListParamsFromQuery("", accountModel.Account{})
	params.AddLeftJoin(accountTypeTableName, fmt.Sprintf("accounts.type_id = %s.id", accountTypeTableName))
	params.AddFilter("userId", []string{targetUID}, list_params.OperatorEq)
	params.AddFilter("accountTypes.currencyCode", []string{currency}, list_params.OperatorEq)

	accounts, err := s.accounts.GetList(params)
	if err != nil {
		return &errors.PrivateError{OriginalError: err}
	}

	if len(accounts) == 0 {
		return errcodes.CreatePublicError(errcodes.CodeCurrencyMismatch)
	}

	return nil
}
