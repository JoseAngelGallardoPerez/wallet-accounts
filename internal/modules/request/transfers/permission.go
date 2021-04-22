package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

// PermissionChecker defines the contract that declares how to check if transfer is allowed
type PermissionChecker interface {
	// Check checks whether transfer is allowed
	Check() error
	// Name returns permission specific name
	Name() string
}

// PermissionCheckers is a list of PermissionChecker
type PermissionCheckers []PermissionChecker

// Check checks the rules one by one, it stops on the first failed rule.
func (r PermissionCheckers) Check() error {
	for _, r := range r {
		if err := r.Check(); err != nil {
			return err
		}
	}
	return nil
}

func (r PermissionCheckers) Name() string {
	return "combined_permissions"
}

// DepositPermission checks whether deposit to a given account is allowed
type DepositPermission struct {
	account *model.Account
}

// NewDepositPermission is DepositPermission constructor
func NewDepositPermission(account *model.Account) *DepositPermission {
	return &DepositPermission{account: account}
}

// Check checks whether rule is satisfied
func (d *DepositPermission) Check() error {
	if d.account.AllowDeposits == nil {
		return errors.Wrapf(
			ErrDepositNotAllowed,
			"account id %d, property 'AllowDeposits' is nil",
			d.account.ID,
		)
	}
	if !*d.account.AllowDeposits {
		return ErrDepositNotAllowed
	}
	return nil
}

func (d *DepositPermission) Name() string {
	return "deposit_allowed"
}

// WithdrawalPermission checks whether whether withdrawal from a given account is allowed
type WithdrawalPermission struct {
	account *model.Account
}

// NewWithdrawalPermission is WithdrawalPermission constructor
func NewWithdrawalPermission(account *model.Account) *WithdrawalPermission {
	return &WithdrawalPermission{account: account}
}

// Check checks whether rule is satisfied
func (w *WithdrawalPermission) Check() error {
	if w.account.AllowWithdrawals == nil {
		return errors.Wrapf(
			ErrWithdrawalNotAllowed,
			"account id %d, property 'AllowWithdrawals' is nil",
			w.account.ID,
		)
	}
	if !*w.account.AllowWithdrawals {
		return ErrWithdrawalNotAllowed
	}
	return nil
}

func (w *WithdrawalPermission) Name() string {
	return "withdrawal_allowed"
}

// AccountActivePermission checks whether account is active
type AccountActivePermission struct {
	account *model.Account
}

// NewAccountActivePermission is NewAccountActivePermission constructor
func NewAccountActivePermission(account *model.Account) *AccountActivePermission {
	return &AccountActivePermission{account: account}
}

func (a *AccountActivePermission) Check() error {
	if a.account.IsActive == nil {
		return errors.Wrapf(
			ErrAccountInactive,
			"account id %d, property 'IsActive' is nil",
			a.account.ID,
		)
	}
	if !*a.account.IsActive {
		return ErrAccountInactive
	}
	return nil
}

func (a *AccountActivePermission) Name() string {
	return "account_active"
}

// PermissionFactory is used in order to define permissions
type PermissionFactory interface {
	CreatePermission(request *requestModel.Request, details types.Details) (PermissionChecker, error)
	WrapContext(db *gorm.DB) PermissionFactory
}

type defaultPermissionFactory struct {
	db                 *gorm.DB
	limitFactory       limit.Factory
	limitStorage       limit.TransactionalStorage
	aggregationService *balance.AggregationService
	logger             log15.Logger
}

func NewDefaultPermissionFactory(
	db *gorm.DB,
	limitStorage limit.Storage,
	limitFactory limit.Factory,
	aggregationService *balance.AggregationService,
	logger log15.Logger,
) PermissionFactory {
	return &defaultPermissionFactory{
		db:                 db,
		limitFactory:       limitFactory,
		limitStorage:       NewLimitStorageDecorator(limitStorage),
		aggregationService: aggregationService,
		logger:             logger,
	}
}

func (d *defaultPermissionFactory) CreatePermission(request *requestModel.Request, details types.Details) (PermissionChecker, error) {
	limitService := limit.NewService(d.limitStorage, d.limitFactory)
	aggregationService := d.aggregationService
	permissions := PermissionCheckers{}

	debitAccounts := make(map[uint64]*model.Account)
	creditAccounts := make(map[uint64]*model.Account)
	for _, detail := range details {
		if detail.Account != nil {
			if detail.IsCredit() {
				creditAccounts[detail.Account.ID] = detail.Account
				continue
			}
			debitAccounts[detail.Account.ID] = detail.Account
		}
	}

	for _, account := range debitAccounts {
		requestedAmount := SimpleAmountable(details.TotalAccountDebit(account.ID).Abs())
		availableAmount := SimpleAmountable(account.AvailableAmount)
		permissions = append(
			permissions,
			NewAccountActivePermission(account),
			NewWithdrawalPermission(account),
			NewSufficientBalancePermission(requestedAmount, availableAmount),
		)
	}
	for _, account := range creditAccounts {
		permissions = append(
			permissions,
			NewAccountActivePermission(account),
			NewDepositPermission(account),
		)
	}

	if LimitMaxTotalBalanceEnabled {
		permissions = append(
			permissions,
			NewMaxBalanceLimit(details, limitService, aggregationService, d.logger),
		)
	}

	if LimitMaxDebitPerTransferEnabled {
		permissions = append(
			permissions,
			NewMaxDebitPerTransfer(details, limitService, aggregationService, d.logger),
		)
	}

	if LimitMaxCreditPerTransferEnabled {
		permissions = append(
			permissions,
			NewMaxCreditPerTransfer(details, limitService, aggregationService, d.logger),
		)
	}

	if LimitMaxTotalDebitPerDayEnabled {
		permissions = append(
			permissions,
			NewMaxTotalDebitPerPeriod(
				details,
				limitService,
				aggregationService,
				LimitMaxTotalDebitPerDay,
				now.BeginningOfDay(),
				now.EndOfDay(),
				d.logger,
			),
		)
	}

	if LimitMaxTotalDebitPerMonthEnabled {
		permissions = append(
			permissions,
			NewMaxTotalDebitPerPeriod(
				details,
				limitService,
				aggregationService,
				LimitMaxTotalDebitPerMonth,
				now.BeginningOfMonth(),
				now.EndOfMonth(),
				d.logger,
			),
		)
	}

	return permissions, nil
}

func (d defaultPermissionFactory) WrapContext(db *gorm.DB) PermissionFactory {
	d.db = db
	d.aggregationService = d.aggregationService.WrapContext(db)
	d.limitStorage = d.limitStorage.WrapContext(db)
	return &d
}
