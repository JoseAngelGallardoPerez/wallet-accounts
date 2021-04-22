package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/repository"
	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
)

type IwtBankAccountService struct {
	repo              *repository.IwtBankAccountRepository
	systemLogsService *system_logs.SystemLogsService
}

func NewIwtBankAccountService(
	repo *repository.IwtBankAccountRepository,
	systemLogsService *system_logs.SystemLogsService,
) *IwtBankAccountService {
	return &IwtBankAccountService{
		repo,
		systemLogsService,
	}
}

// Create creates new iwt bank account
func (a *IwtBankAccountService) Create(
	account *model.IwtBankAccountModel,
	user *userpb.User,
) (*model.IwtBankAccountModel, error) {
	res, err := a.repo.Create(account)
	if err != nil {
		return nil, err
	}

	a.systemLogsService.LogCreateIwtBankDetailsAsync(res, user.UID)

	return res, nil
}

// Create creates new iwt bank account
func (a *IwtBankAccountService) Delete(
	account *model.IwtBankAccountModel,
	user *userpb.User,
) error {
	err := a.repo.Delete(account)
	if err != nil {
		return err
	}

	a.systemLogsService.LogDeleteIwtBankDetailsAsync(account, user.UID)

	return nil
}

// Create creates new iwt bank account
func (a *IwtBankAccountService) Update(
	account *model.IwtBankAccountModel,
	data *model.IwtBankAccountModel,
	user *userpb.User,
) (*model.IwtBankAccountModel, error) {
	old, _ := a.repo.FindByID(account.ID) // @TODO: implement clone instead
	res, err := a.repo.Update(account, data)
	if err != nil {
		return nil, err
	}

	a.systemLogsService.LogModifyIwtBankDetailsAsync(old, res, user.UID)

	return res, nil
}
