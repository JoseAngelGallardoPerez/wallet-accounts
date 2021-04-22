package transfers

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-accounts/internal/transfer/fee"
)

// CreditAccountInput defines required CA input data
type CreditAccountInput interface {
	Account() *model.Account
	ApplyIWTFee() bool
	DebitFromRevenueAccount() bool
	RevenueAccount() *model.RevenueAccountModel
	FeeParams() *fee.TransferFeeParams
}

type creditAccountInput struct {
	account          *model.Account
	applyIWTFee      bool
	debitFromRevenue bool
	revenueAccount   *model.RevenueAccountModel
	feeParams        *fee.TransferFeeParams
}

func (c *creditAccountInput) FeeParams() *fee.TransferFeeParams {
	return c.feeParams
}

// NewCreditAccountInput wraps given arguments into the container that implements CreditAccountInput interface
func NewCreditAccountInput(
	account *model.Account,
	applyIWTFee bool,
	debitFromRevenue bool,
	revenueAccount *model.RevenueAccountModel,
	feeParams *fee.TransferFeeParams,
) CreditAccountInput {
	return &creditAccountInput{
		account:          account,
		applyIWTFee:      applyIWTFee,
		debitFromRevenue: debitFromRevenue,
		revenueAccount:   revenueAccount,
		feeParams:        feeParams,
	}
}

func (c *creditAccountInput) Account() *model.Account {
	return c.account
}

func (c *creditAccountInput) ApplyIWTFee() bool {
	return c.applyIWTFee
}

func (c *creditAccountInput) DebitFromRevenueAccount() bool {
	return c.debitFromRevenue
}

func (c *creditAccountInput) RevenueAccount() *model.RevenueAccountModel {
	return c.revenueAccount
}
