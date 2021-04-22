package transfers

import accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"

type DaInput interface {
	SourceAccount() (*accountModel.Account, error)
	RevenueAccount() (*accountModel.RevenueAccountModel, error)
	CreditToRevenueAccount() (bool, error)
	AllowNegativeBalance() (bool, error)
}

type daInput struct {
	sourceAccount        *accountModel.Account
	revenueAccount       *accountModel.RevenueAccountModel
	creditToRevenue      bool
	allowNegativeBalance bool
}

func NewDaInput(
	sourceAccount *accountModel.Account,
	revenueAccount *accountModel.RevenueAccountModel,
	creditToRevenue bool,
	allowNegativeBalance bool,
) DaInput {
	return &daInput{
		sourceAccount:        sourceAccount,
		revenueAccount:       revenueAccount,
		creditToRevenue:      creditToRevenue,
		allowNegativeBalance: allowNegativeBalance,
	}
}

func (d *daInput) SourceAccount() (*accountModel.Account, error) {
	return d.sourceAccount, nil
}

func (d *daInput) RevenueAccount() (*accountModel.RevenueAccountModel, error) {
	return d.revenueAccount, nil
}

func (d *daInput) CreditToRevenueAccount() (bool, error) {
	return d.creditToRevenue, nil
}

func (d *daInput) AllowNegativeBalance() (bool, error) {
	return d.allowNegativeBalance, nil
}
