package service

import (
	"github.com/Confialink/wallet-accounts/internal/modules/country/model"
	"github.com/Confialink/wallet-accounts/internal/modules/country/repository"
)

type CountryService struct {
	repo *repository.CountryRepository
}

func NewCountryService(repo *repository.CountryRepository) *CountryService {
	return &CountryService{repo}
}

// FindById find country by id
func (s *CountryService) FindById(id *uint) (*model.Country, error) {
	return s.repo.FindById(id)
}

// FindAll find all countries
func (s *CountryService) FindAll() ([]*model.Country, error) {
	return s.repo.FindAll()
}
