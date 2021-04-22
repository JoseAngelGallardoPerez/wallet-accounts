package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/model"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const defaultLimit = 30

type CardTypeCategoryRepository struct {
	db     *gorm.DB
	logger log15.Logger
}

func NewCardTypeCategoryRepository(db *gorm.DB, logger log15.Logger) *CardTypeCategoryRepository {
	return &CardTypeCategoryRepository{db: db, logger: logger.New("Repository", "CardTypeCategoryRepository")}
}

func (self *CardTypeCategoryRepository) GetList() ([]*model.CardTypeCategory, error) {
	var cardTypeCategories []*model.CardTypeCategory

	query := self.db.Limit(defaultLimit)

	if err := query.Find(&cardTypeCategories).Error; err != nil {
		return cardTypeCategories, err
	}

	return cardTypeCategories, nil
}

func (repo *CardTypeCategoryRepository) FindByID(id uint32) (*model.CardTypeCategory, error) {
	var cardTypeCategory model.CardTypeCategory
	cardTypeCategory.Id = &id
	if err := repo.db.First(&cardTypeCategory).Error; err != nil {
		return nil, err
	}
	return &cardTypeCategory, nil
}
