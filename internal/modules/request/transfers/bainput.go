package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/conv"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type BetweenAccountsInput interface {
	SourceAccount() (*model.Account, error)
	DestinationAccount() (*model.Account, error)
	RevenueAccount() (*model.RevenueAccountModel, error)
	ExchangeMarginPercent() (decimal.Decimal, error)
	TransferFeeParams() (*fee.TransferFeeParams, error)
}

type BaInputCache struct {
	SourceAccount         *model.Account
	DestinationAccount    *model.Account
	RevenueAccount        *model.RevenueAccountModel
	ExchangeMarginPercent *decimal.Decimal
	TransferFeeParams     *fee.TransferFeeParams
}

type betweenAccountsInput struct {
	sourceAccount         *model.Account
	destinationAccount    *model.Account
	revenueAccount        *model.RevenueAccountModel
	exchangeMarginPercent decimal.Decimal
	transferFeeParams     *fee.TransferFeeParams
}

func (b *betweenAccountsInput) SourceAccount() (*model.Account, error) {
	return b.sourceAccount, nil
}

func (b *betweenAccountsInput) DestinationAccount() (*model.Account, error) {
	return b.destinationAccount, nil
}

func (b *betweenAccountsInput) RevenueAccount() (*model.RevenueAccountModel, error) {
	return b.revenueAccount, nil
}

func (b *betweenAccountsInput) ExchangeMarginPercent() (decimal.Decimal, error) {
	return b.exchangeMarginPercent, nil
}

func (b *betweenAccountsInput) TransferFeeParams() (*fee.TransferFeeParams, error) {
	return b.transferFeeParams, nil
}

func NewBetweenAccountsInput(
	sourceAccount *model.Account,
	destinationAccount *model.Account,
	revenueAccount *model.RevenueAccountModel,
	exchangeMarginPercent decimal.Decimal,
	transferFeeParams *fee.TransferFeeParams,
) BetweenAccountsInput {
	return &betweenAccountsInput{
		sourceAccount:         sourceAccount,
		destinationAccount:    destinationAccount,
		revenueAccount:        revenueAccount,
		exchangeMarginPercent: exchangeMarginPercent,
		transferFeeParams:     transferFeeParams,
	}
}

type dbBetweenAccountsInput struct {
	db      *gorm.DB
	request *requestModel.Request

	cache BaInputCache
}

func NewDbBetweenAccountsInput(
	db *gorm.DB,
	request *requestModel.Request,
	cache *BaInputCache,
) BetweenAccountsInput {
	input := &dbBetweenAccountsInput{
		db:      db,
		request: request,
	}
	if cache != nil {
		input.cache = *cache
	}
	return input
}

func (d *dbBetweenAccountsInput) SourceAccount() (*model.Account, error) {
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

func (d *dbBetweenAccountsInput) DestinationAccount() (*model.Account, error) {
	if d.cache.DestinationAccount != nil {
		return d.cache.DestinationAccount, nil
	}
	param, _ := d.request.GetInput().Get("destinationAccountId")
	accountId := conv.Int64FromInterface(param)
	if accountId == 0 {
		return nil, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "destinationAccountId" field`,
		)
	}
	account, err := getAccountWithTypeForUpdateById(d.db, accountId)
	if err != nil {
		return nil, err
	}
	d.cache.DestinationAccount = account
	return account, nil
}

func (d *dbBetweenAccountsInput) RevenueAccount() (*model.RevenueAccountModel, error) {
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

func (d *dbBetweenAccountsInput) ExchangeMarginPercent() (result decimal.Decimal, err error) {
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

func (d *dbBetweenAccountsInput) TransferFeeParams() (result *fee.TransferFeeParams, err error) {
	if d.cache.TransferFeeParams != nil {
		return d.cache.TransferFeeParams, nil
	}
	result, err = transferFeeParamsFromRequest(d.request)
	if err != nil {
		d.cache.TransferFeeParams = result
	}

	return
}
