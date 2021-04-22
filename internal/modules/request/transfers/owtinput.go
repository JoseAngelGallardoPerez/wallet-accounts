package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/conv"
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type OWTInput interface {
	SourceAccount() (*accountModel.Account, error)
	RevenueAccount() (*accountModel.RevenueAccountModel, error)
	ExchangeMarginPercent() (decimal.Decimal, error)
	TransferFeeParams() (*fee.TransferFeeParams, error)
	BeneficiaryCustomerAccountName() (string, error)
	RefMessage() (string, error)
}

type owtInput struct {
	sourceAccount                  *accountModel.Account
	revenueAccount                 *accountModel.RevenueAccountModel
	exchangeMarginPercent          decimal.Decimal
	transferFeeParams              *fee.TransferFeeParams
	beneficiaryCustomerAccountName string
	refMessage                     string
}

type OWTInputCache struct {
	SourceAccount                  *accountModel.Account
	RevenueAccount                 *accountModel.RevenueAccountModel
	ExchangeMarginPercent          *decimal.Decimal
	TransferFeeParams              *fee.TransferFeeParams
	BeneficiaryCustomerAccountName *string
	RefMessage                     *string
}

func NewOwtInput(
	sourceAccount *accountModel.Account,
	revenueAccount *accountModel.RevenueAccountModel,
	exchangeMarginPercent decimal.Decimal,
	transferFeeParams *fee.TransferFeeParams,
	beneficiaryCustomerAccountName string,
	refMessage string,
) OWTInput {
	return &owtInput{
		sourceAccount:                  sourceAccount,
		revenueAccount:                 revenueAccount,
		exchangeMarginPercent:          exchangeMarginPercent,
		transferFeeParams:              transferFeeParams,
		beneficiaryCustomerAccountName: beneficiaryCustomerAccountName,
		refMessage:                     refMessage,
	}
}

func (o *owtInput) SourceAccount() (*accountModel.Account, error) {
	return o.sourceAccount, nil
}

func (o *owtInput) RevenueAccount() (*accountModel.RevenueAccountModel, error) {
	return o.revenueAccount, nil
}

func (o *owtInput) ExchangeMarginPercent() (decimal.Decimal, error) {
	return o.exchangeMarginPercent, nil
}

func (o *owtInput) TransferFeeParams() (*fee.TransferFeeParams, error) {
	return o.transferFeeParams, nil
}

func (o *owtInput) BeneficiaryCustomerAccountName() (string, error) {
	return o.beneficiaryCustomerAccountName, nil
}

func (o *owtInput) RefMessage() (string, error) {
	return o.refMessage, nil
}

type dbOWTInput struct {
	db      *gorm.DB
	request *requestModel.Request

	cache *OWTInputCache
}

func NewDbOWTInput(db *gorm.DB, request *requestModel.Request, cache *OWTInputCache) OWTInput {
	if cache == nil {
		cache = &OWTInputCache{}
	}
	return &dbOWTInput{db: db, request: request, cache: cache}
}

func (d *dbOWTInput) SourceAccount() (*accountModel.Account, error) {
	if d.cache.SourceAccount != nil {
		return d.cache.SourceAccount, nil
	}
	param, _ := d.request.GetInput().Get("sourceAccountId")
	accountId := conv.Int64FromInterface(param)
	if accountId == 0 {
		return nil, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "sourceAccountId" field`,
		)
	}
	account, err := getAccountWithTypeForUpdateById(d.db, accountId)
	if err != nil {
		return nil, err
	}
	d.cache.SourceAccount = account
	return account, nil
}

func (d *dbOWTInput) RevenueAccount() (*accountModel.RevenueAccountModel, error) {
	if d.cache.RevenueAccount != nil {
		return d.cache.RevenueAccount, nil
	}
	param, _ := d.request.GetInput().Get("revenueAccountId")
	accountId := conv.Int64FromInterface(param)
	if accountId == 0 {
		return nil, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "revenueAccountId" field`,
		)
	}
	account, err := getRevenueAccountForUpdateById(d.db, accountId)
	if err != nil {
		return nil, err
	}
	d.cache.RevenueAccount = account
	return account, nil
}

func (d *dbOWTInput) ExchangeMarginPercent() (result decimal.Decimal, err error) {
	if d.cache.ExchangeMarginPercent != nil {
		return *d.cache.ExchangeMarginPercent, nil
	}
	input := d.request.GetInput()
	param, ok := input["exchangeMarginPercent"]
	if !ok {
		return result, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "exchangeMarginPercent" field`,
		)
	}
	result, err = decimalFromInterface(param, "exchangeMarginPercent")
	if err != nil {
		return
	}
	d.cache.ExchangeMarginPercent = &result
	return *d.cache.ExchangeMarginPercent, nil
}

func (d *dbOWTInput) TransferFeeParams() (result *fee.TransferFeeParams, err error) {
	if d.cache.TransferFeeParams != nil {
		return d.cache.TransferFeeParams, nil
	}
	result, err = transferFeeParamsFromRequest(d.request)
	if err != nil {
		d.cache.TransferFeeParams = result
	}

	return
}

func (d *dbOWTInput) BeneficiaryCustomerAccountName() (result string, err error) {
	if d.cache.BeneficiaryCustomerAccountName != nil {
		return *d.cache.BeneficiaryCustomerAccountName, nil
	}
	input := d.request.GetInput()
	param, ok := input["beneficiaryCustomerAccountName"]
	if !ok {
		return result, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "beneficiaryCustomerAccountName" field`,
		)
	}
	result = param.(string)
	return
}

func (d *dbOWTInput) RefMessage() (result string, err error) {
	if d.cache.RefMessage != nil {
		return *d.cache.RefMessage, nil
	}
	input := d.request.GetInput()
	param, ok := input["refMessage"]
	if !ok {
		return result, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "refMessage" field`,
		)
	}
	result = param.(string)
	return
}
