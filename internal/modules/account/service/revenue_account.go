package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type RevenueAccountService struct {
	repo *repository.RevenueAccountRepository
}

func NewRevenueAccountService(repo *repository.RevenueAccountRepository) *RevenueAccountService {
	return &RevenueAccountService{repo}
}

func (r *RevenueAccountService) FindOrCreateDefaultByCurrencyCode(currencyCode string, db *gorm.DB) (*model.RevenueAccountModel, error) {
	repo := r.repo.WrapContext(db)
	account, err := repo.FindDefaultByCurrencyCode(currencyCode)
	if nil == account {
		account = &model.RevenueAccountModel{
			RevenueAccountPublic: model.RevenueAccountPublic{
				Balance:      decimal.NewFromFloat(0),
				CurrencyCode: currencyCode,
				IsDefault:    true,
			},
		}
		err = repo.Create(account)
	}
	return account, err
}
