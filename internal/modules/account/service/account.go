package service

import (
	"encoding/json"
	"fmt"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	userpb "github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	accountTypeRepo "github.com/Confialink/wallet-accounts/internal/modules/account-type/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account/form"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	accountRepo "github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	system_logs "github.com/Confialink/wallet-accounts/internal/modules/system-logs"
)

type AccountService struct {
	db                *gorm.DB
	accountRepo       *accountRepo.AccountRepository
	accountTypeRepo   *accountTypeRepo.AccountTypeRepository
	systemLogsService *system_logs.SystemLogsService
}

func NewAccountService(
	db *gorm.DB,
	accountRepo *accountRepo.AccountRepository,
	accountTypeRepo *accountTypeRepo.AccountTypeRepository,
	systemLogsService *system_logs.SystemLogsService,
) *AccountService {
	return &AccountService{
		db,
		accountRepo,
		accountTypeRepo,
		systemLogsService,
	}
}

func (a *AccountService) GetList(listParams *list_params.ListParams) (
	[]*model.Account, error) {
	return a.accountRepo.GetList(listParams)
}

func (a *AccountService) GetCount(listParams *list_params.ListParams) (
	int64, error) {
	return a.accountRepo.GetCount(listParams)
}

func (a *AccountService) FindByID(id uint64, preloads ...string) (*model.Account, error) {
	return a.accountRepo.FindByID(id, preloads...)
}

// Create creates new account
func (a *AccountService) Create(
	account *model.Account,
	user *userpb.User,
) (*model.Account, errors.TypedError) {
	accRepo := a.accountRepo.WrapContext(a.db)

	typeRecord, _ := a.accountTypeRepo.WrapContext(a.db).FindByID(account.TypeID)
	if nil == typeRecord {
		return nil, errcodes.CreatePublicError(errcodes.CodeAccountTypeNotFound, fmt.Sprintf("account type #%d not found", account.TypeID))
	}

	if nil != account.InterestAccountId {
		interestAccountRecord, _ := accRepo.FindByID(*account.InterestAccountId)
		if nil == interestAccountRecord {
			return nil, errcodes.CreatePublicError(errcodes.CodeAccountNotFound, fmt.Sprintf("account #%d is not exist", account.InterestAccountId))
		}
	}

	// available amount must be the same value as credit limit
	// in order to allow negative balance transactions (overdraft)
	if typeRecord.CreditLimitAmount != nil {
		account.AvailableAmount = typeRecord.CreditLimitAmount.Add(account.Balance)
	}

	res, err := accRepo.Create(account)
	if err != nil {
		return nil, &errors.PrivateError{Message: err.Error()}
	}

	a.systemLogsService.LogCreateAccountAsync(res, user.UID)

	return res, nil
}

func (a *AccountService) Update(
	account *model.Account, editable *model.AccountEditable, user *userpb.User,
) (*model.Account, errors.TypedError) {
	old, _ := a.accountRepo.FindByID(account.ID) // @TODO: implement clone instead

	editableJSON, err := json.Marshal(editable)
	if nil != err {
		pvtErr := errors.PrivateError{Message: "can't marshal json"}
		pvtErr.AddLogPair("error", err.Error())
		return nil, &pvtErr
	}

	err = json.Unmarshal(editableJSON, &account)
	if err != nil {
		pvtErr := errors.PrivateError{Message: "Can't unmarshal json"}
		pvtErr.AddLogPair("err", err)
		return nil, &pvtErr
	}

	updatedAccount, err := a.accountRepo.Update(account)

	if nil != err {
		requestData, _ := json.Marshal(account)
		pvtErr := &errors.PrivateError{Message: "can't update account"}
		pvtErr.AddLogPair("error", err)
		pvtErr.AddLogPair("account id", account.ID)
		pvtErr.AddLogPair("request data", requestData)
		return nil, pvtErr
	}

	a.systemLogsService.LogModifyAccountAsync(old, updatedAccount, user.UID)

	return updatedAccount, nil
}

// BulkCreate creates list of accounts
func (a *AccountService) BulkCreate(accounts []*model.Account) ([]*model.Account, error) {
	accRepo := a.accountRepo.WrapContext(a.db)

	for _, account := range accounts {
		accountRecord, _ := accRepo.FindByNumber(account.Number)
		if nil != accountRecord {
			return nil, errcodes.CreatePublicError(errcodes.CodeDuplicateAccountNumber, fmt.Sprintf("account number #%s already exists", account.Number))
		}

		typeRecord, _ := a.accountTypeRepo.WrapContext(a.db).FindByID(account.TypeID)
		if nil == typeRecord {
			return nil, errcodes.CreatePublicError(errcodes.CodeAccountTypeNotFound, fmt.Sprintf("account type #%d not found", account.TypeID))
		}

		if nil != account.InterestAccountId {
			interestAccountRecord, _ := accRepo.FindByID(*account.InterestAccountId)
			if nil == interestAccountRecord {
				return nil, errcodes.CreatePublicError(errcodes.CodeAccountNotFound, fmt.Sprintf("account #%d is not exist", account.InterestAccountId))
			}
		}

		// available amount must be the same value as credit limit
		// in order to allow negative balance transactions (overdraft)
		if typeRecord.CreditLimitAmount != nil {
			account.AvailableAmount = typeRecord.CreditLimitAmount.Add(account.Balance)
		}
	}

	res, err := accRepo.BulkCreate(accounts)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (a *AccountService) CreateAccountAsWallet(
	currency string,
	uid string,
) (*model.Account, error) {
	accRepo := a.accountRepo.WrapContext(a.db)

	typeRecord, err := a.accountTypeRepo.WrapContext(a.db).FindByCurrencyCode(currency)
	if err != nil {
		return nil, err
	}
	if typeRecord == nil {
		return nil, fmt.Errorf("account type for currency %s not found", currency)
	}

	f := form.GenerateNumber{}
	number := a.GenerateAccountNumberWithPrefix(f.Prefix)

	account := model.Account{
		AccountPublic: model.AccountPublic{
			Number:           number,
			TypeID:           typeRecord.ID,
			UserId:           uid,
			IsActive:         pointer.ToBool(true),
			AllowWithdrawals: pointer.ToBool(true),
			AllowDeposits:    pointer.ToBool(true),
		},
	}

	res, err := accRepo.Create(&account)
	if err != nil {
		return nil, &errors.PrivateError{Message: err.Error()}
	}

	a.systemLogsService.LogCreateAccountAsync(res, uid)

	return res, nil
}

func (a *AccountService) CountByAccountTypeId(id uint64) (uint64, error) {
	return a.accountRepo.CountByAccountTypeId(id)
}

func (a *AccountService) GenerateAccountNumberWithPrefix(prefix *string) string {
	return a.accountRepo.GenerateAccountNumberWithPrefix(prefix)
}

func (a *AccountService) UserIncludes(records []interface{}) error {
	accounts := make([]*model.Account, len(records))
	for i, v := range records {
		accounts[i] = v.(*model.Account)
	}
	return a.accountRepo.FillUsers(accounts)
}

func (a AccountService) WrapContext(db *gorm.DB) *AccountService {
	a.db = db
	return &a
}
