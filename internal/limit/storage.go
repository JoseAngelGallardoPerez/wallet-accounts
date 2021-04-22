package limit

import (
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

// Storage defines how to store and fetch limits
type Storage interface {
	// Save saves new limit with the given identifier
	Save(value Value, identifier Identifier) error
	// Update updates available amount on existing limit by its identifier
	Update(value Value, identifier Identifier) error
	// Find retrieves all limits that much the given identifier parameters
	// must return ErrNotFound if no records is found
	Find(identifier Identifier) ([]Model, error)
	// Delete deletes all limits that much the given identifier parameters
	Delete(identifier Identifier) error
}

// TransactionalStorage is used in order to pass running db transaction
type TransactionalStorage interface {
	Storage
	WrapContext(db *gorm.DB) TransactionalStorage
}

type dbModel struct {
	ID           int64
	Amount       *decimal.Decimal
	CurrencyCode string
	Name         string
	Entity       string
	EntityId     string
}

type StorageGORM struct {
	db *gorm.DB
}

func NewStorageGORM(db *gorm.DB) Storage {
	return &StorageGORM{db: db}
}

func (s *StorageGORM) Save(value Value, identifier Identifier) error {
	code, amount := currencyAndAmount(value)
	return s.db.
		Exec(
			"INSERT INTO `limits`(`currency_code`, `amount`, `name`, `entity`, `entity_id`) VALUES (?, ?, ?, ?, ?)",
			code,
			amount,
			identifier.Name,
			identifier.Entity,
			identifier.EntityId,
		).Error
}

func (s *StorageGORM) Update(value Value, identifier Identifier) error {
	code, amount := currencyAndAmount(value)
	return s.db.
		Exec(
			"UPDATE `limits` SET `currency_code` = ?, `amount` = ? WHERE `name` = ? AND `entity` = ? AND `entity_id` = ?",
			code,
			amount,
			identifier.Name,
			identifier.Entity,
			identifier.EntityId,
		).Error
}

func (s *StorageGORM) Find(identifier Identifier) ([]Model, error) {
	var found []dbModel
	db := s.buildWhere(identifier)

	if err := db.Find(&found).Error; err != nil {
		return nil, err
	}
	result := make([]Model, len(found))
	for i, m := range found {
		var value Value
		if m.Amount == nil {
			value = NoLimit()
		} else {
			value = Val(*m.Amount, m.CurrencyCode)
		}
		result[i] = Model{
			Identifier: Identifier{
				Name:     m.Name,
				Entity:   m.Entity,
				EntityId: m.EntityId,
			},
			Value: value,
		}
	}
	return result, nil
}

func (s *StorageGORM) Delete(identifier Identifier) error {
	return s.buildWhere(identifier).Delete(nil).Error
}

// WrapContext make a copy of the storage using passed db
func (s StorageGORM) WrapContext(db *gorm.DB) TransactionalStorage {
	s.db = db
	return &s
}

func (s *StorageGORM) buildWhere(identifier Identifier) *gorm.DB {
	db := s.db.Table("limits")
	if identifier.Name != "" {
		db = db.Where("`name` = ?", identifier.Name)
	}
	if identifier.Entity != "" {
		db = db.Where("`entity` = ?", identifier.Entity)
	}
	if identifier.EntityId != "" {
		db = db.Where("`entity_id` = ?", identifier.EntityId)
	}
	return db
}

func currencyAndAmount(value Value) (*string, interface{}) {
	if value.NoLimit() {
		return nil, nil
	}
	actual := value.CurrencyAmount()
	code, amount := actual.CurrencyCode(), actual.Amount()

	return &code, &amount
}
