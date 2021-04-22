package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/conv"
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type CFTInput interface {
	SourceAccount() (*model.Account, error)
	DestinationCard() (*cardModel.Card, error)
	RevenueAccount() (*model.RevenueAccountModel, error)
	ExchangeMarginPercent() (decimal.Decimal, error)
	TransferFeeParams() (*fee.TransferFeeParams, error)
}

type CFTInputCache struct {
	SourceAccount         *model.Account
	DestinationCard       *cardModel.Card
	RevenueAccount        *model.RevenueAccountModel
	ExchangeMarginPercent *decimal.Decimal
	TransferFeeParams     *fee.TransferFeeParams
}

type cftInput struct {
	sourceAccount         *model.Account
	destinationCard       *cardModel.Card
	revenueAccount        *model.RevenueAccountModel
	exchangeMarginPercent decimal.Decimal
	transferFeeParams     *fee.TransferFeeParams
}

func (c *cftInput) SourceAccount() (*model.Account, error) {
	return c.sourceAccount, nil
}

func (c *cftInput) DestinationCard() (*cardModel.Card, error) {
	return c.destinationCard, nil
}

func (c *cftInput) RevenueAccount() (*model.RevenueAccountModel, error) {
	return c.revenueAccount, nil
}

func (c *cftInput) ExchangeMarginPercent() (decimal.Decimal, error) {
	return c.exchangeMarginPercent, nil
}

func (c *cftInput) TransferFeeParams() (*fee.TransferFeeParams, error) {
	return c.transferFeeParams, nil
}

func NewCFTInput(
	sourceAccount *model.Account,
	destinationCard *cardModel.Card,
	revenueAccount *model.RevenueAccountModel,
	exchangeMarginPercent decimal.Decimal,
	transferFeeParams *fee.TransferFeeParams,
) CFTInput {
	return &cftInput{
		sourceAccount:         sourceAccount,
		destinationCard:       destinationCard,
		revenueAccount:        revenueAccount,
		exchangeMarginPercent: exchangeMarginPercent,
		transferFeeParams:     transferFeeParams,
	}
}

type dbCFTInput struct {
	db      *gorm.DB
	request *requestModel.Request

	cache CFTInputCache
}

func NewDbCFTInput(
	db *gorm.DB,
	request *requestModel.Request,
	cache *CFTInputCache,
) CFTInput {
	input := &dbCFTInput{
		db:      db,
		request: request,
	}
	if cache != nil {
		input.cache = *cache
	}
	return input
}

func (c *dbCFTInput) SourceAccount() (*model.Account, error) {
	if c.cache.SourceAccount != nil {
		return c.cache.SourceAccount, nil
	}
	param, _ := c.request.GetInput().Get("sourceAccountId")
	accountId := conv.Int64FromInterface(param)
	if accountId == 0 {
		return nil, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "sourceAccountId" field`,
		)
	}
	account, err := getAccountWithTypeForUpdateById(c.db, accountId)
	if err != nil {
		return nil, err
	}
	c.cache.SourceAccount = account
	return account, nil
}

func (c *dbCFTInput) DestinationCard() (*cardModel.Card, error) {
	if c.cache.DestinationCard != nil {
		return c.cache.DestinationCard, nil
	}
	param, _ := c.request.GetInput().Get("destinationCardId")
	cardId := conv.Int64FromInterface(param)
	if cardId == 0 {
		return nil, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "destinationCardId" field`,
		)
	}
	card, err := getCardWithTypeForUpdateById(c.db, cardId)
	if err != nil {
		return nil, err
	}
	c.cache.DestinationCard = card
	return card, nil
}

func (c *dbCFTInput) RevenueAccount() (*model.RevenueAccountModel, error) {
	if c.cache.RevenueAccount != nil {
		return c.cache.RevenueAccount, nil
	}
	param, _ := c.request.GetInput().Get("revenueAccountId")
	accountId := conv.Int64FromInterface(param)
	if accountId == 0 {
		return nil, errors.Wrap(
			ErrMissingInputData,
			`request input must contain "revenueAccountId" field`,
		)
	}
	account, err := getRevenueAccountForUpdateById(c.db, accountId)
	if err != nil {
		return nil, err
	}
	c.cache.RevenueAccount = account
	return account, nil
}

func (c *dbCFTInput) ExchangeMarginPercent() (result decimal.Decimal, err error) {
	if c.cache.ExchangeMarginPercent != nil {
		return *c.cache.ExchangeMarginPercent, nil
	}
	input := c.request.GetInput()
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
	c.cache.ExchangeMarginPercent = &result
	return *c.cache.ExchangeMarginPercent, nil
}

func (c *dbCFTInput) TransferFeeParams() (result *fee.TransferFeeParams, err error) {
	if c.cache.TransferFeeParams != nil {
		return c.cache.TransferFeeParams, nil
	}
	result, err = transferFeeParamsFromRequest(c.request)
	if err != nil {
		c.cache.TransferFeeParams = result
	}

	return
}
