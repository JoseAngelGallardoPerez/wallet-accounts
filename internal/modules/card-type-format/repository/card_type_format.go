package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/model"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
)

const defaultLimit = 30

type CardTypeFormatRepository struct {
	db     *gorm.DB
	logger log15.Logger
}

func NewCardTypeFormatRepository(db *gorm.DB, logger log15.Logger) *CardTypeFormatRepository {
	return &CardTypeFormatRepository{db: db, logger: logger.New("Repository", "CardTypeFormatRepository")}
}

func (self *CardTypeFormatRepository) GetList() ([]*model.CardTypeFormat, error) {
	var cardTypeFormats []*model.CardTypeFormat

	query := self.db.Limit(defaultLimit)

	if err := query.Find(&cardTypeFormats).Error; err != nil {
		return cardTypeFormats, err
	}

	return cardTypeFormats, nil
}

func (repo *CardTypeFormatRepository) FindByID(id uint32) (*model.CardTypeFormat, error) {
	var cardTypeFormat model.CardTypeFormat
	cardTypeFormat.Id = &id
	if err := repo.db.First(&cardTypeFormat).Error; err != nil {
		return nil, err
	}
	return &cardTypeFormat, nil
}
