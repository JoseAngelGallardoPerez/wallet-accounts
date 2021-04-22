package form

import (
	"github.com/Confialink/wallet-accounts/internal/modules/fee/model"
	"github.com/shopspring/decimal"
)

type TransferFeeParameters struct {
	CurrencyCode *string `json:"currencyCode" binding:"required"`
	Base         *string `json:"base" binding:"omitempty,decimal,decimalGT=0"`
	Min          *string `json:"min" binding:"omitempty,decimal,decimalGT=0"`
	Percent      *string `json:"percent" binding:"omitempty,decimal,decimalGT=0"`
	Max          *string `json:"max" binding:"omitempty,decimal,decimalGT=0"`
	Delete       *bool   `json:"delete"`
}

type TransferFeeParametersList []*TransferFeeParameters

func (t *TransferFeeParameters) ToModel() (*model.TransferFeeParameters, error) {
	parametersModel := &model.TransferFeeParameters{
		CurrencyCode: t.CurrencyCode,
	}

	base, err := t.DecimalBase()
	if err != nil {
		return nil, err
	}
	parametersModel.Base = base

	min, err := t.DecimalMin()
	if err != nil {
		return nil, err
	}
	parametersModel.Min = min

	percent, err := t.DecimalPercent()
	if err != nil {
		return nil, err
	}
	parametersModel.Percent = percent

	max, err := t.DecimalMax()
	if err != nil {
		return nil, err
	}
	parametersModel.Max = max

	return parametersModel, nil
}

func (t *TransferFeeParameters) DecimalBase() (*decimal.Decimal, error) {
	if t.Base == nil {
		return nil, nil
	}
	value, err := decimal.NewFromString(*t.Base)
	if err != nil {
		return nil, err
	}

	return &value, nil
}

func (t *TransferFeeParameters) DecimalPercent() (*decimal.Decimal, error) {
	if t.Percent == nil {
		return nil, nil
	}
	value, err := decimal.NewFromString(*t.Percent)
	if err != nil {
		return nil, err
	}

	return &value, nil
}

func (t *TransferFeeParameters) DecimalMin() (*decimal.Decimal, error) {
	if t.Min == nil {
		return nil, nil
	}
	value, err := decimal.NewFromString(*t.Min)
	if err != nil {
		return nil, err
	}

	return &value, nil
}

func (t *TransferFeeParameters) DecimalMax() (*decimal.Decimal, error) {
	if t.Max == nil {
		return nil, nil
	}
	value, err := decimal.NewFromString(*t.Max)
	if err != nil {
		return nil, err
	}

	return &value, nil
}
