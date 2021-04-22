package service

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type AccountTypeService struct {
	repo   *repository.AccountTypeRepository
	logger log15.Logger
}

func NewAccountTypeService(
	repo *repository.AccountTypeRepository,
	logger log15.Logger,
) *AccountTypeService {
	return &AccountTypeService{
		repo:   repo,
		logger: logger.New("Service", "AccountTypeService"),
	}
}

// checkNameAndCurrencyCode checked account type by currency code and name
func (a *AccountTypeService) checkNameAndCurrencyCode(name, currencyCode string) error {
	_, err := a.repo.FindByNameAndCurrencyCode(name, currencyCode)

	if err == nil {
		return errcodes.CreatePublicError(errcodes.AccountTypeNameIsDuplicated, "Name and currency are already in use.")
	}

	if gorm.IsRecordNotFoundError(err) {
		return nil
	}

	return errors.Wrap(err, "failed to create new account type")
}

func (a *AccountTypeService) Update(account *model.AccountType) (*model.AccountType, error) {
	oldAccType, err := a.repo.FindByID(account.ID)
	if err != nil {
		return nil, err
	}

	err = a.checkNameAndCurrencyCode(account.Name, account.CurrencyCode)

	if (err != nil) && ((account.Name != oldAccType.Name) || (account.CurrencyCode != oldAccType.CurrencyCode)) {
		return nil, err
	}

	updatedAccountType, err := a.repo.Update(account)

	if err != nil {
		return nil, err
	}

	return updatedAccountType, nil
}

func (a *AccountTypeService) Create(account *model.AccountType) (*model.AccountType, error) {
	err := a.checkNameAndCurrencyCode(account.Name, account.CurrencyCode)
	if err != nil {
		return nil, err
	}

	createAccountType, err := a.repo.Create(account)

	if err != nil {
		return nil, err
	}

	return createAccountType, nil
}
