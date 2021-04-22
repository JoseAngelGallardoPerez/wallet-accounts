package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/balance/model"
	"github.com/jinzhu/gorm"
)

type Snapshot struct {
	db *gorm.DB
}

func NewSnapshot(db *gorm.DB) *Snapshot {
	return &Snapshot{db: db}
}

func (s *Snapshot) Save(snapshot *model.Snapshot) error {
	return s.db.Save(snapshot).Error
}

func (s *Snapshot) Create(snapshot *model.Snapshot) error {
	return s.db.Create(snapshot).Error
}

func (s *Snapshot) Updates(snapshot *model.Snapshot) error {
	return s.db.Model(snapshot).Updates(snapshot).Error
}

func (s *Snapshot) FindOrInitByRequestIdAndBalanceId(requestId, balanceId uint64) (*model.Snapshot, error) {
	result := &model.Snapshot{}
	err := s.db.
		Model(result).
		FirstOrInit(result, "request_id = ? and balance_id = ?", requestId, balanceId).
		Error

	return result, err
}

func (s Snapshot) WrapContext(db *gorm.DB) *Snapshot {
	s.db = db
	return &s
}
