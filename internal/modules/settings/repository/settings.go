package repository

import (
	"errors"
	"fmt"
	"log"

	"github.com/Confialink/wallet-accounts/internal/modules/settings/form"
	"github.com/Confialink/wallet-accounts/internal/modules/settings/model"
	"github.com/jinzhu/gorm"
)

type Settings struct {
	db *gorm.DB
}

func NewSettings(db *gorm.DB) *Settings {
	return &Settings{db: db}
}

func (s *Settings) FirstByName(name string) (*model.Settings, error) {
	setting := &model.Settings{}
	return setting, s.db.Where("name = ?", name).FirstOrInit(setting).Error
}

func (s *Settings) Update(name, value string) (*model.Settings, error) {
	setting, err := s.FirstByName(name)
	if nil != err {
		return nil, err
	}
	if !setting.IsExist() {
		return nil, errors.New(fmt.Sprintf("setting %s does not exist", name))
	}

	setting.Value = value
	return setting, s.db.Save(setting).Error
}

func (s *Settings) MassUpdate(pairs []*form.KeyValue) error {
	for _, pair := range pairs {
		if _, err := s.Update(pair.Key, pair.Value); err != nil {
			return err
		}
	}
	return nil
}

func (s Settings) WrapContext(tx *gorm.DB) *Settings {
	s.db = tx
	return &s
}

func (s *Settings) GetAll() ([]*model.Settings, error) {
	result := make([]*model.Settings, 0, 16)
	err := s.db.Find(&result).Error
	if nil != err {
		log.Println("settings repository GetAll: ", err)
	}
	return result, err
}
