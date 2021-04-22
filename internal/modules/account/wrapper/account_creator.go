package wrapper

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/Confialink/wallet-pkg-utils/pointer"
	"github.com/Confialink/wallet-pkg-utils/value"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/modules/account/form"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-accounts/internal/modules/account/service"
	"github.com/Confialink/wallet-accounts/internal/modules/request"
	requestForm "github.com/Confialink/wallet-accounts/internal/modules/request/form"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
)

const ApplyIwtFeeWhenCreateNewAccount = "apply_iwt_fee_when_create_new_account"
const DebitFromRevenueWhenCreateNewAccount = "debit_from_revenue_when_create_new_account"

type AccountCreator struct {
	db                *gorm.DB
	accountService    *service.AccountService
	requestCreator    *request.Creator
	settingsService   *settings.Service
	accountRepository *repository.AccountRepository
}

func NewAccountCreator(
	db *gorm.DB,
	accountService *service.AccountService,
	requestCreator *request.Creator,
	settingsService *settings.Service,
	accountRepository *repository.AccountRepository,
) *AccountCreator {
	return &AccountCreator{
		db,
		accountService,
		requestCreator,
		settingsService,
		accountRepository,
	}
}

// CreateAccountWithRequest creates new account and request with transaction if balance is not zero
func (c *AccountCreator) CreateAccountWithRequest(form *form.Account, user *users.User) (*model.Account, errors.TypedError) {
	tx := c.db.Begin()
	accountService := c.accountService.WrapContext(tx)

	allowDeposits := value.FromBool(form.AllowDeposits)
	// if initial balance is requested and deposits are not allowed we have to temporary enable it
	if !form.InitialBalance.IsZero() && !allowDeposits {
		allowDeposits = true
	}
	account := model.Account{
		AccountPublic: model.AccountPublic{
			Number:            form.Number,
			TypeID:            form.TypeId,
			UserId:            form.UserId,
			Description:       form.Description,
			IsActive:          form.IsActive,
			AllowWithdrawals:  form.AllowWithdrawals,
			AllowDeposits:     &allowDeposits,
			MaturityDate:      form.MaturityDate,
			PayoutDay:         form.PayoutDay,
			InterestAccountId: form.InterestAccountId,
			InitialBalance:    &form.InitialBalance,
		},
	}

	res, typedErr := accountService.Create(&account, user)
	if typedErr != nil {
		tx.Rollback()
		return nil, typedErr
	}

	if !form.InitialBalance.IsZero() {
		applyIwtFee, _ := c.settingsService.Bool(ApplyIwtFeeWhenCreateNewAccount)
		debitFromRevenueAccount, _ := c.settingsService.Bool(DebitFromRevenueWhenCreateNewAccount)
		reqForm := requestForm.CA{
			AccountId:               res.ID,
			Amount:                  form.InitialBalance.String(),
			ApplyIwtFee:             &applyIwtFee,
			DebitFromRevenueAccount: &debitFromRevenueAccount,
			Description:             "New Account",
		}
		_, err := c.requestCreator.CreateCARequest(&reqForm, user, tx, true)
		if err != nil {
			tx.Rollback()
			return nil, errcodes.ConvertToTyped(err)
		}

		accRepoTx := c.accountRepository.WrapContext(tx)
		// reload with balance filled
		res, err = accRepoTx.FindByID(res.ID)
		if err != nil {
			tx.Rollback()
			return nil, &errors.PrivateError{Message: err.Error()}
		}

		// if we temporary enabled deposits now it must be disabled
		if allowDeposits && !value.FromBool(form.AllowDeposits) {
			err = accRepoTx.Updates(
				&model.Account{
					AccountPrivate: model.AccountPrivate{ID: account.ID},
					// nil value is ignored so it must be not nil pointer to bool
					AccountPublic: model.AccountPublic{AllowDeposits: pointer.ToBool(false)},
				},
			)
			if err != nil {
				tx.Rollback()
				return nil, &errors.PrivateError{Message: err.Error()}
			}
			res.AllowDeposits = form.AllowDeposits
		}
	}

	tx.Commit()

	return res, nil
}

// BulkCreateAccountWithRequest creates new accounts and requests with transactions if balance is not zero
func (c *AccountCreator) BulkCreateAccountWithRequest(forms []*form.Account, user *users.User) ([]*model.Account, error) {
	var accounts []*model.Account
	for _, f := range forms {
		res, err := c.CreateAccountWithRequest(f, user)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, res)
	}

	return accounts, nil
}
