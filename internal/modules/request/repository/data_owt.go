package repository

import (
	"github.com/jinzhu/gorm"

	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
)

type DataOwt struct {
	db *gorm.DB
}

func NewDataOwt(db *gorm.DB) *DataOwt {
	return &DataOwt{db: db}
}

func (d *DataOwt) Create(owt *model.DataOwt) error {
	return d.db.Create(owt).Error
}
func (d *DataOwt) FindByRequestId(requestId uint64) (*model.DataOwt, error) {
	data := &model.DataOwt{}
	err := d.
		db.
		Preload("BankDetails").
		Preload("BankDetails.Country").
		Preload("BeneficiaryCustomer").
		Preload("IntermediaryBankDetails").
		Preload("IntermediaryBankDetails.Country").
		Find(data, "request_id = ?", requestId).
		Error

	return data, err
}

func (d DataOwt) WrapContext(db *gorm.DB) *DataOwt {
	d.db = db
	return &d
}
