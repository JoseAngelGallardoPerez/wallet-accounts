package repository

import (
	"net/url"

	"github.com/Confialink/wallet-accounts/internal/modules/payment-method/model"
	"github.com/jinzhu/gorm"
)

type PaymentMethodRepositoryInterface interface {
	FindByParams(params url.Values) ([]*model.PayoutMethodModel, error)
}

// Repository is user repository for CRUD operations.
type PaymentMethodRepository struct {
	db *gorm.DB
}

// NewRepository creates new repository
func NewPaymentMethodRepository(db *gorm.DB) PaymentMethodRepositoryInterface {
	return &PaymentMethodRepository{db}
}

// FindByParams retrieve the list of messages
func (repo *PaymentMethodRepository) FindByParams(params url.Values) ([]*model.PayoutMethodModel, error) {
	var paymentMethods []*model.PayoutMethodModel

	order := "id asc" //p.Order

	if err := repo.db.
		Order(order).
		Find(&paymentMethods).
		Error; err != nil {
		return nil, err
	}

	return paymentMethods, nil
}
