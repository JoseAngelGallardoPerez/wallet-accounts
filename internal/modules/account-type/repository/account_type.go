package repository

import (
	"net/url"
	"strconv"

	"github.com/Confialink/wallet-accounts/internal/modules/account-type/event"
	"github.com/olebedev/emitter"

	"github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/jinzhu/gorm"
)

// Repository is user repository for CRUD operations.
type AccountTypeRepository struct {
	db           *gorm.DB
	eventEmitter *emitter.Emitter
}

// NewRepository creates new repository
func NewAccountTypeRepository(db *gorm.DB, emitter *emitter.Emitter) *AccountTypeRepository {
	return &AccountTypeRepository{db: db, eventEmitter: emitter}
}

// FindByParams retrieve the list of messages
func (repo *AccountTypeRepository) FindByParams(params url.Values) ([]*model.AccountType, error) {
	var accounts []*model.AccountType

	order := "id asc" //p.Order

	allowedSortFields := []string{"name", "credit_annual_interest_rate", "credit_charge_period_id"}
	allowedSortDirs := []string{"asc", "desc"}

	sortField := params.Get("sort[field]")
	sortDir := params.Get("sort[dir]")

	if len(sortField) > 0 &&
		len(sortDir) > 0 &&
		repo.sliceContains(allowedSortFields, sortField) &&
		repo.sliceContains(allowedSortDirs, sortDir) {
		order = sortField + " " + sortDir
	}

	query := repo.db

	if len(params.Get("name")) > 0 {
		query = query.Where("name LIKE ?", "%"+params.Get("name")+"%")
	}

	query = repo.paginate(query, params)

	if err := query.
		Order(order).
		Find(&accounts).
		Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

// CountByParams retrieve the list of messages
func (repo *AccountTypeRepository) CountByParams(params url.Values) (*int64, error) {
	var accounts []*model.AccountType
	var count int64

	query := repo.db

	if len(params.Get("name")) > 0 {
		query = query.Where("name LIKE ?", "%"+params.Get("name")+"%")
	}

	query = repo.paginate(query, params)

	if err := query.
		Find(&accounts).
		Count(&count).
		Error; err != nil {
		return nil, err
	}

	return &count, nil
}

// FindByID find user by id
func (repo *AccountTypeRepository) FindByID(id uint64) (*model.AccountType, error) {
	var accountType model.AccountType
	accountType.ID = id
	if err := repo.db.
		Preload("DepositPayoutMethod").
		Preload("DepositPayoutPeriod").
		Preload("CreditPayoutMethod").
		Preload("CreditChargePeriod").
		First(&accountType).
		Error; err != nil {
		return nil, err
	}
	return &accountType, nil
}

// FindByCurrencyCode find account type by currency code
func (repo *AccountTypeRepository) FindByCurrencyCode(currencyCode string) (*model.AccountType, error) {
	var accountType model.AccountType
	if err := repo.db.Where("currency_code = ?", currencyCode).First(&accountType).Error; err != nil {
		return nil, err
	}
	return &accountType, nil
}

// FindByNameAndCurrencyCode finds account type by currency code and name
func (repo *AccountTypeRepository) FindByNameAndCurrencyCode(name, currencyCode string) (*model.AccountType, error) {
	var accountType model.AccountType
	if err := repo.db.Where("name = ? AND currency_code = ?", name, currencyCode).First(&accountType).Error; err != nil {
		return nil, err
	}
	return &accountType, nil
}

// Create creates new account
func (repo *AccountTypeRepository) Create(account *model.AccountType) (*model.AccountType, error) {
	if err := repo.db.Create(account).Error; err != nil {
		return nil, err
	}
	return repo.FindByID(account.ID)
}

// Update updates an existing account
func (repo *AccountTypeRepository) Update(account *model.AccountType) (*model.AccountType, error) {
	oldAccType := &model.AccountType{}
	err := repo.db.FirstOrInit(oldAccType, account.ID).Error
	if err != nil {
		return nil, err
	}

	tx := repo.db.Begin()
	// we use Save() instead Update() because we need to save empty fields
	if err := tx.Model(account).Save(account).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	evtContext := &event.ContextAccountTypeUpdated{
		OldAccountType: oldAccType,
		NewAccountType: account,
		DbTransaction:  tx,
	}
	// wait while event is processing
	<-repo.eventEmitter.Emit(event.AccountTypeUpdated, evtContext)

	if evtContext.Error != nil {
		tx.Rollback()
		return nil, evtContext.Error
	}

	err = tx.Commit().Error
	if err != nil {
		return nil, err
	}

	return repo.FindByID(account.ID)
}

// Delete delete an existing account
func (repo *AccountTypeRepository) Delete(account *model.AccountType) error {
	if err := repo.db.Delete(account).Error; err != nil {
		return err
	}
	return nil
}

// paginate check if query parameters limit, after and offset are set
// and applies them to query builder
// argument usedLimit will be set the same value as for limit
// it is used in order to determine whether there are more records exist
// limit always increments by 1
func (repo *AccountTypeRepository) paginate(query *gorm.DB, params url.Values) *gorm.DB {
	limit, err := strconv.ParseUint(params.Get("limit"), 10, 32)
	if nil == err {
		if limit > 100 {
			limit = 100
		}

	} else {
		limit = 15
	}

	query = query.Limit(uint(limit + 1))
	offset, err := strconv.ParseUint(params.Get("offset"), 10, 32)
	if nil == err {
		query = query.Offset(uint(offset))
	}

	return query
}

// GetList returns records from passed ListParams
func (self *AccountTypeRepository) GetList(params *list_params.ListParams) (
	[]*model.AccountType, error,
) {
	var accountTypes []*model.AccountType

	str, arguments := params.GetWhereCondition()
	query := self.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	query = query.Limit(params.GetLimit())
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	if err := query.Find(&accountTypes).Error; err != nil {
		return accountTypes, err
	}

	interfaceAccountTypess := make([]interface{}, len(accountTypes))
	for i, accountTypePtr := range accountTypes {
		interfaceAccountTypess[i] = accountTypePtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceAccountTypess); err != nil {
			return accountTypes, err
		}
	}

	return accountTypes, nil
}

func (self AccountTypeRepository) WrapContext(db *gorm.DB) *AccountTypeRepository {
	self.db = db
	return &self
}

func (self *AccountTypeRepository) GetListCount(params *list_params.ListParams) (uint64, error) {
	var count uint64

	query := self.db.Joins(params.GetJoinCondition())
	str, arguments := params.GetWhereCondition()
	query = query.Where(str, arguments...)

	if err := query.Model(&model.AccountType{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

// sliceContains checks string exists in slice
func (repo *AccountTypeRepository) sliceContains(slice []string, string string) bool {
	for _, el := range slice {
		if el == string {
			return true
		}
	}
	return false
}
