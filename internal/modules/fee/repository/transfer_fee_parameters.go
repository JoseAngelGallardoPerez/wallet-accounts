package repository

import (
	"github.com/Confialink/wallet-accounts/internal/modules/fee/model"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
)

type TransferFeeParameters struct {
	db *gorm.DB
}

func NewTransferFeeParameters(db *gorm.DB) *TransferFeeParameters {
	return &TransferFeeParameters{db: db}
}

func (t *TransferFeeParameters) CreateUpdate(parameters *model.TransferFeeParameters) error {
	where := model.TransferFeeParameters{TransferFeeId: parameters.TransferFeeId, CurrencyCode: parameters.CurrencyCode}
	// map b/c fields might be updated with null values
	assign := map[string]*decimal.Decimal{
		"base":    parameters.Base,
		"min":     parameters.Min,
		"percent": parameters.Percent,
		"max":     parameters.Max,
	}

	return t.db.Where(where).Assign(assign).FirstOrCreate(&parameters).Error
}

func (t *TransferFeeParameters) Delete(parameters *model.TransferFeeParameters) error {
	where := model.TransferFeeParameters{Id: parameters.Id}
	if where.Id == nil {
		where = model.TransferFeeParameters{TransferFeeId: parameters.TransferFeeId, CurrencyCode: parameters.CurrencyCode}
	}
	return t.db.Delete(parameters, where).Error
}

func (t *TransferFeeParameters) GetAllByTransferFeeId(transferFeeId uint64) (params []*model.TransferFeeParameters, err error) {
	err = t.db.Model(&model.TransferFeeParameters{}).Find(&params, "transfer_fee_id = ?", transferFeeId).Error
	return
}

func (t *TransferFeeParameters) FindByTransferFeeIdAndCurrencyCode(transferFeeId uint64, currencyCode string) (*model.TransferFeeParameters, error) {
	resultModel := &model.TransferFeeParameters{}
	err := t.db.FirstOrInit(resultModel, "transfer_fee_id = ? AND currency_code = ?", transferFeeId, currencyCode).Error
	return resultModel, err
}

func (copy TransferFeeParameters) WrapContext(db *gorm.DB) *TransferFeeParameters {
	copy.db = db
	return &copy
}
