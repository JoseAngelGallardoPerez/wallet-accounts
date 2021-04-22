package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/country/model"
	"github.com/jinzhu/gorm"
)

// Repository is user repository for CRUD operations.
type CountryRepository struct {
	db *gorm.DB
}

// NewRepository creates new repository
func NewCountryRepository(db *gorm.DB) *CountryRepository {
	return &CountryRepository{db}
}

// FindById find country by id
func (r *CountryRepository) FindById(id *uint) (*model.Country, error) {
	var country model.Country
	country.Id = id
	if err := r.db.First(&country).Error; err != nil {
		return nil, err
	}
	return &country, nil
}

// FindAll find all countries
func (r *CountryRepository) FindAll() (
	[]*model.Country, error,
) {
	var countries []*model.Country

	if err := r.db.Find(&countries).Error; err != nil {
		return countries, err
	}

	return countries, nil
}
