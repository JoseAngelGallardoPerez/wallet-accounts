package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type-category/repository"
	"github.com/inconshreveable/log15"
)

type CardTypeCategoryServiceInterface interface {
	GetList() ([]*model.CardTypeCategory, error)
}

type cardTypeCategoryService struct {
	repo   *repository.CardTypeCategoryRepository
	logger log15.Logger
}

func NewCardTypeCategoryService(
	repo *repository.CardTypeCategoryRepository,
	logger log15.Logger,

) CardTypeCategoryServiceInterface {
	return &cardTypeCategoryService{repo: repo, logger: logger.New("Service", "CardTypeCategoryService")}
}

func (self *cardTypeCategoryService) GetList() (
	[]*model.CardTypeCategory, error,
) {
	return self.repo.GetList()
}
