package repository

import (
	"net/url"

	"github.com/Confialink/wallet-accounts/internal/modules/payment-period/model"
	"github.com/jinzhu/gorm"
)

type PaymentPeriodRepositoryInterface interface {
	FindByParams(params url.Values) ([]*model.PaymentPeriodModel, error)
}

// Repository is user repository for CRUD operations.
type PaymentPeriodRepository struct {
	db *gorm.DB
}

// NewRepository creates new repository
func NewPaymentPeriodRepository(db *gorm.DB) PaymentPeriodRepositoryInterface {
	return &PaymentPeriodRepository{db}
}

// FindByParams retrieve the list of messages
func (repo *PaymentPeriodRepository) FindByParams(params url.Values) ([]*model.PaymentPeriodModel, error) {
	var paymentPeriods []*model.PaymentPeriodModel

	order := "id asc" //p.Order

	if err := repo.db.
		Order(order).
		Find(&paymentPeriods).
		Error; err != nil {
		return nil, err
	}

	return paymentPeriods, nil
}
