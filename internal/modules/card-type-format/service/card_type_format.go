package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-format/repository"
	"github.com/inconshreveable/log15"
)

type CardTypeFormatServiceInterface interface {
	GetList() ([]*model.CardTypeFormat, error)
}

type cardTypeFormatService struct {
	repo   *repository.CardTypeFormatRepository
	logger log15.Logger
}

func NewCardTypeFormatService(
	repo *repository.CardTypeFormatRepository,
	logger log15.Logger,

) CardTypeFormatServiceInterface {
	return &cardTypeFormatService{repo: repo, logger: logger.New("Service", "CardTypeFormatService")}
}

func (self *cardTypeFormatService) GetList() (
	[]*model.CardTypeFormat, error,
) {
	return self.repo.GetList()
}
