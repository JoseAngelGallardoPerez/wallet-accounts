package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance/model"
	"github.com/jinzhu/gorm"
)

type Type struct {
	db *gorm.DB

	typesMap map[string]*model.Type
}

func NewType(db *gorm.DB) *Type {
	return &Type{db: db, typesMap: make(map[string]*model.Type)}
}

func (t *Type) FindByName(name string) (*model.Type, error) {
	if t, ok := t.typesMap[name]; ok {
		return &*t, nil
	}

	result := &model.Type{}
	err := t.db.Model(result).Where("name = ?", name).Find(result).Error
	if err != nil {
		return nil, err
	}

	t.typesMap[name] = result

	return &*result, nil
}

func (t Type) WrapContext(db *gorm.DB) *Type {
	t.db = db
	return &t
}
