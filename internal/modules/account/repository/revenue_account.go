package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	listParams "github.com/Confialink/wallet-pkg-list_params"
	"github.com/jinzhu/gorm"
)

// Repository is user repository for CRUD operations.
type RevenueAccountRepository struct {
	db *gorm.DB
}

// NewRepository creates new repository
func NewRevenueAccountRepository(db *gorm.DB) *RevenueAccountRepository {
	return &RevenueAccountRepository{db: db}
}

// Create creates new revenue account
func (repo *RevenueAccountRepository) Create(account *model.RevenueAccountModel) error {
	if err := repo.db.Create(account).Error; err != nil {
		return err
	}
	return nil
}

// Updates updates only not default values of an existing revenue account
func (r *RevenueAccountRepository) Updates(account *model.RevenueAccountModel) error {
	return r.db.Model(account).Updates(account).Error
}

// FindByID find revenue account by id
func (r *RevenueAccountRepository) FindByID(id uint64) (*model.RevenueAccountModel, error) {
	var account model.RevenueAccountModel
	account.ID = id
	if err := r.db.First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// FindDefaultByCurrencyCode find revenue account by default flag and currency id
func (r *RevenueAccountRepository) FindDefaultByCurrencyCode(currencyCode string) (*model.RevenueAccountModel, error) {
	account := &model.RevenueAccountModel{}
	account.CurrencyCode = currencyCode
	account.IsDefault = true
	if err := r.db.First(account, account).Error; err != nil {
		return nil, err
	}
	return account, nil
}

// GetList returns records from passed ListParams
func (r *RevenueAccountRepository) GetList(params *listParams.ListParams) (
	[]*model.RevenueAccountModel, error,
) {
	var accounts []*model.RevenueAccountModel

	query := r.db.
		Limit(params.GetLimit()).
		Offset(params.GetOffset())

	if err := query.Find(&accounts).Error; err != nil {
		return accounts, err
	}

	return accounts, nil
}

func (r *RevenueAccountRepository) GetListCount() (uint64, error) {
	var count uint64

	if err := r.db.Model(&model.RevenueAccountModel{}).Count(&count).Error; err != nil {
		return count, err
	}

	return count, nil
}

func (r *RevenueAccountRepository) WrapContext(db *gorm.DB) *RevenueAccountRepository {
	return NewRevenueAccountRepository(db)
}
