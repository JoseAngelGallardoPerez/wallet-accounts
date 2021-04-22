package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/limit"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// limitDefaultValuesStorageDecorator is used in order to provide default values for certain limits
type limitDefaultValuesStorageDecorator struct {
	storage limit.Storage
}

func NewLimitStorageDecorator(storage limit.Storage) limit.TransactionalStorage {
	return &limitDefaultValuesStorageDecorator{storage: storage}
}

// Save directly calls underlying storage
func (l *limitDefaultValuesStorageDecorator) Save(value limit.Value, identifier limit.Identifier) error {
	return l.storage.Save(value, identifier)
}

// Update directly calls underlying storage
func (l *limitDefaultValuesStorageDecorator) Update(value limit.Value, identifier limit.Identifier) error {
	return l.storage.Update(value, identifier)
}

// Find decorates underlying storage method.
// In case if default storage does not find limit values for limits with names:
// "max_debit_per_transfer", "max_total_balance", "max_total_debit_per_day", "max_total_debit_per_month"
// then the decorator storage provides default values based on the corresponding constants
func (l *limitDefaultValuesStorageDecorator) Find(identifier limit.Identifier) ([]limit.Model, error) {
	result, err := l.storage.Find(identifier)
	if err != nil && errors.Cause(err) != limit.ErrNotFound {
		return result, err
	}
	notFound := errors.Cause(err) == limit.ErrNotFound || len(result) == 0
	if notFound {
		var value limit.Value
		switch identifier.Name {
		case LimitMaxTotalBalance:
			value = limit.Val(
				decimal.NewFromFloat(LimitMaxTotalBalanceDefaultAmount),
				LimitMaxTotalBalanceDefaultCurrency,
			)
		case LimitMaxDebitPerTransfer:
			value = limit.Val(
				decimal.NewFromFloat(LimitMaxDebitPerTransferDefaultAmount),
				LimitMaxDebitPerTransferDefaultCurrency,
			)
		case LimitMaxTotalDebitPerDay:
			value = limit.Val(
				decimal.NewFromFloat(LimitMaxTotalDebitPerDayDefaultAmount),
				LimitMaxTotalDebitPerDayDefaultCurrency,
			)
		case LimitMaxTotalDebitPerMonth:
			value = limit.Val(
				decimal.NewFromFloat(LimitMaxTotalDebitPerMonthDefaultAmount),
				LimitMaxTotalDebitPerMonthDefaultCurrency,
			)
		case LimitMaxCreditPerTransfer:
			value = limit.Val(
				decimal.NewFromFloat(LimitMaxCreditPerTransferDefaultAmount),
				LimitMaxCreditPerTransferDefaultCurrency,
			)
		}
		if value != nil {
			err = nil
			result = []limit.Model{
				{
					Identifier: identifier,
					Value:      value,
				},
			}
		}
	}
	return result, err
}

// Delete directly calls underlying storage
func (l *limitDefaultValuesStorageDecorator) Delete(identifier limit.Identifier) error {
	return l.storage.Delete(identifier)
}

// WrapContext checks whether underlying storage is transactional and calls its wrapper method if true
func (l *limitDefaultValuesStorageDecorator) WrapContext(db *gorm.DB) limit.TransactionalStorage {
	if storage, ok := l.storage.(limit.TransactionalStorage); ok {
		l.storage = storage.WrapContext(db)
	}
	return l
}
