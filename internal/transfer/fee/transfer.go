package fee

import (
	"github.com/Confialink/wallet-accounts/internal/transfer"
	"github.com/shopspring/decimal"
)

// TransferFeeParams used in order to define transfer fee parameters
type TransferFeeParams struct {
	// Base is fixed amount
	Base decimal.Decimal `json:"base"`
	// Percent is percentage of a given amount
	Percent decimal.Decimal `json:"percent"`
	// Min is only used when Percent is not zero. This value limits the minimum transfer fee amount.
	Min decimal.Decimal `json:"min"`
	// Max is only used when Percent is not zero. This value limits the maximum transfer fee amount.
	Max decimal.Decimal `json:"max"`
}

// NewDebitTransferFeeAction is TransferFee constructor
func NewDebitTransferFeeAction(debitable transfer.Debitable, fromAmount transfer.CurrencyAmount, params TransferFeeParams) (*transfer.DebitAction, error) {
	return transfer.NewDebitAction(debitable, NewTransferFeeAmount(params, fromAmount))
}

// TransferFeeAmount is used in order to calculate transfer fee
type TransferFeeAmount struct {
	params     TransferFeeParams
	fromAmount transfer.CurrencyAmount
}

// NewTransferFeeAmount is TransferFeeAmount constructor
func NewTransferFeeAmount(params TransferFeeParams, fromAmount transfer.CurrencyAmount) *TransferFeeAmount {
	return &TransferFeeAmount{params: params, fromAmount: fromAmount}
}

// Currency returns transfer fee currency
func (t *TransferFeeAmount) Currency() transfer.Currency {
	return t.fromAmount.Currency()
}

// Amount calculates transfer fee
func (t *TransferFeeAmount) Amount() decimal.Decimal {
	zero := decimal.New(0, 1)
	totalFee := t.params.Base
	if t.params.Percent.GreaterThan(zero) {
		percentFee := t.fromAmount.Amount().Mul(t.params.Percent.Div(decimal.New(100, 0)))
		if t.params.Min.GreaterThan(zero) && percentFee.LessThan(t.params.Min) {
			percentFee = t.params.Min
		}
		if t.params.Max.GreaterThan(zero) && percentFee.GreaterThan(t.params.Max) {
			percentFee = t.params.Max
		}
		totalFee = totalFee.Add(percentFee)
	}
	return totalFee
}
