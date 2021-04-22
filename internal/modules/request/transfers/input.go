package transfers

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func transferFeeParamsFromRequest(request *model.Request) (result *fee.TransferFeeParams, err error) {
	input := request.GetInput()
	param, ok := input["transferFeeParams"]
	if !ok {
		return result, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "transferFeeParams" field`,
		)
	}
	switch param := param.(type) {
	case nil:
		return nil, nil
	case map[string]interface{}:
		sm := param
		result = &fee.TransferFeeParams{}
		baseParam, ok := sm["base"]
		if !ok {
			return nil, errors.Wrap(
				ErrMissingInputData,
				`request input must contain "transferFeeParams.base" field`,
			)
		}
		base, err := decimalFromInterface(baseParam, "transferFeeParams.base")
		if err != nil {
			return result, err
		}
		result.Base = base

		percentParam, ok := sm["percent"]
		if !ok {
			return nil, errors.Wrap(
				ErrMissingInputData,
				`request input must contain "transferFeeParams.percent" field`,
			)
		}
		percent, err := decimalFromInterface(percentParam, "transferFeeParams.percent")
		if err != nil {
			return result, err
		}
		result.Percent = percent

		minParam, ok := sm["min"]
		if !ok {
			return nil, errors.Wrap(
				ErrMissingInputData,
				`request input must contain "transferFeeParams.min" field`,
			)
		}
		min, err := decimalFromInterface(minParam, "transferFeeParams.min")
		if err != nil {
			return result, err
		}
		result.Min = min

		maxParam, ok := sm["max"]
		if !ok {
			return nil, errors.Wrap(
				ErrMissingInputData,
				`request input must contain "transferFeeParams.max" field`,
			)
		}
		max, err := decimalFromInterface(maxParam, "transferFeeParams.max")
		if err != nil {
			return result, err
		}
		result.Max = max

	case fee.TransferFeeParams:
		tfp := param
		result = &tfp
	case *fee.TransferFeeParams:
		result = param
	default:
		return result, errors.Wrapf(
			ErrMissingInputData,
			`parameter "transferFeeParams" has wrong type, expected types are "map[string]interface{}", "fee.TransferFeeParams", "*fee.TransferFeeParams" but got "%T"`,
			param,
		)
	}
	return
}

func getAccountWithTypeForUpdateById(db *gorm.DB, accountId int64) (*accountModel.Account, error) {
	account := &accountModel.Account{}
	err := db.
		Preload("Type").
		Raw("SELECT * FROM `accounts` WHERE `accounts`.`id` = ? FOR UPDATE", accountId).
		Find(account).
		Error
	return account, err
}

func getCardWithTypeForUpdateById(db *gorm.DB, cardId int64) (*cardModel.Card, error) {
	card := &cardModel.Card{}
	err := db.
		Preload("CardType").
		Raw("SELECT * FROM `cards` WHERE `cards`.`id` = ? FOR UPDATE", cardId).
		Find(card).
		Error
	return card, err
}

func getRevenueAccountForUpdateById(db *gorm.DB, accountId int64) (*accountModel.RevenueAccountModel, error) {
	account := &accountModel.RevenueAccountModel{}
	err := db.
		Raw("SELECT * FROM `revenue_accounts` WHERE `revenue_accounts`.`id` = ? FOR UPDATE", accountId).
		Find(account).
		Error
	return account, err
}

func decimalFromInterface(param interface{}, name string) (result decimal.Decimal, err error) {
	switch param := param.(type) {
	case string:
		result, err = decimal.NewFromString(param)
	case float64:
		result = decimal.NewFromFloat(param)
	case float32:
		result = decimal.NewFromFloat32(param)
	case decimal.Decimal:
		result = param
	case *decimal.Decimal:
		result = *param
	case nil:
		result = decimal.NewFromInt(0)
	default:
		err = ErrMissingInputData
	}
	if err != nil {
		return decimal.Decimal{}, errors.Wrapf(
			err,
			`parameter %s has wrong type, expected types are "string", "float32", "float64", "nil", "decimal.Decimal", "*decimal.Decimal" but got "%T"`,
			name,
			param,
		)
	}
	return
}
