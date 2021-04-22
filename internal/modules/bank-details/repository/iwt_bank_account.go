package repository

import (
	"net/url"
	"strconv"

	"github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	"github.com/Confialink/wallet-pkg-list_params"
	"github.com/jinzhu/gorm"
)

// IwtBankAccountRepository repository for Itw bank accounts
type IwtBankAccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates new repository
func NewAccountRepository(db *gorm.DB) *IwtBankAccountRepository {
	return &IwtBankAccountRepository{db}
}

// GetList returns records from passed ListParams
func (repo *IwtBankAccountRepository) GetList(params *list_params.ListParams) (
	[]*model.IwtBankAccountModel, error,
) {
	var items []*model.IwtBankAccountModel

	str, arguments := params.GetWhereCondition()
	query := repo.db.Where(str, arguments...)

	query = query.Order(params.GetOrderByString())

	query = query.Limit(params.GetLimit())
	query = query.Offset(params.GetOffset())

	query = query.Joins(params.GetJoinCondition())

	for _, preloadName := range params.GetPreloads() {
		query = query.Preload(preloadName)
	}

	if err := query.Find(&items).Error; err != nil {
		return items, err
	}

	interfaceIwtBankAccounts := make([]interface{}, len(items))
	for i, itemPtr := range items {
		interfaceIwtBankAccounts[i] = itemPtr
	}
	for _, customIncludesFunc := range params.GetCustomIncludesFunctions() {
		if err := customIncludesFunc(interfaceIwtBankAccounts); err != nil {
			return items, err
		}
	}

	return items, nil
}

// GetListCount returns count of iwt bank accounts
// receive *list_params.ListParams
func (repo *IwtBankAccountRepository) GetListCount(params *list_params.ListParams,
) (uint64, error) {
	var count uint64

	query := repo.db.Joins(params.GetJoinCondition())
	str, arguments := params.GetWhereCondition()
	query = query.Where(str, arguments...)

	if err := query.Model(&model.IwtBankAccountModel{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

// FindByParams retrieve the list
func (repo *IwtBankAccountRepository) FindByParams(params url.Values) ([]*model.IwtBankAccountModel, error) {
	var accounts []*model.IwtBankAccountModel

	order := "id asc" //p.Order

	allowedSortFields := []string{"name"}
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

	query = repo.paginate(query, params)

	if err := query.
		Order(order).
		Preload("BeneficiaryBankDetails").
		Preload("BeneficiaryCustomer").
		Preload("IntermediaryBankDetails").
		Preload("BeneficiaryBankDetails.Country").
		Preload("IntermediaryBankDetails.Country").
		Find(&accounts).
		Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

// CountByParams retrieve the list of messages
func (repo *IwtBankAccountRepository) CountByParams(params url.Values) (*int64, error) {
	var accounts []*model.IwtBankAccountModel
	var count int64

	query := repo.db

	if err := query.
		Find(&accounts).
		Count(&count).
		Error; err != nil {
		return nil, err
	}

	return &count, nil
}

// FindByID find account by id
func (repo *IwtBankAccountRepository) FindByID(id uint64) (*model.IwtBankAccountModel, error) {
	var account model.IwtBankAccountModel
	account.ID = id
	if err := repo.db.
		Preload("BeneficiaryBankDetails").
		Preload("BeneficiaryCustomer").
		Preload("IntermediaryBankDetails").
		Preload("BeneficiaryBankDetails.Country").
		Preload("IntermediaryBankDetails.Country").
		First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// FindByCurrencyCode find iwt accounts by currency code
func (repo *IwtBankAccountRepository) FindByCurrencyCode(code string) ([]*model.IwtBankAccountModel, error) {
	var accounts []*model.IwtBankAccountModel
	if err := repo.db.
		Where("currency_code = ?", code).
		Preload("BeneficiaryBankDetails").
		Preload("BeneficiaryCustomer").
		Preload("IntermediaryBankDetails").
		Preload("BeneficiaryBankDetails.Country").
		Preload("IntermediaryBankDetails.Country").
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// FindEnabledByCurrencyCode find enabled iwt accounts by currency code
func (repo *IwtBankAccountRepository) FindEnabledByCurrencyCode(code string) ([]*model.IwtBankAccountModel, error) {
	var accounts []*model.IwtBankAccountModel
	if err := repo.db.
		Where("currency_code = ? AND is_iwt_enabled = TRUE", code).
		Preload("BeneficiaryBankDetails").
		Preload("BeneficiaryCustomer").
		Preload("IntermediaryBankDetails").
		Preload("BeneficiaryBankDetails.Country").
		Preload("IntermediaryBankDetails.Country").
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// Create creates new account
func (repo *IwtBankAccountRepository) Create(account *model.IwtBankAccountModel) (*model.IwtBankAccountModel, error) {
	if err := repo.db.Create(account).Error; err != nil {
		return nil, err
	}
	return repo.FindByID(account.ID)
}

// Update updates an existing account
func (repo *IwtBankAccountRepository) Update(
	account *model.IwtBankAccountModel,
	data *model.IwtBankAccountModel,
) (*model.IwtBankAccountModel, error) {
	db := repo.db.Model(account)
	if nil == data.IntermediaryBankDetails {
		db.Association("IntermediaryBankDetails").Clear()
	}
	if err := db.Updates(data).Error; err != nil {
		return nil, err
	}
	return repo.FindByID(account.ID)
}

// Delete delete an existing account
func (repo *IwtBankAccountRepository) Delete(account *model.IwtBankAccountModel) error {
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
func (repo *IwtBankAccountRepository) paginate(query *gorm.DB, params url.Values) *gorm.DB {
	limit, err := strconv.ParseUint(params.Get("limit"), 10, 32)
	if nil == err {
		if limit > 100 {
			limit = 100
		}

	} else {
		limit = 15
	}

	query = query.Limit(uint(limit))
	offset, err := strconv.ParseUint(params.Get("offset"), 10, 32)
	if nil == err {
		query = query.Offset(uint(offset))
	}

	return query
}

// sliceContains checks string exists in slice
func (repo *IwtBankAccountRepository) sliceContains(slice []string, string string) bool {
	for _, el := range slice {
		if el == string {
			return true
		}
	}
	return false
}
