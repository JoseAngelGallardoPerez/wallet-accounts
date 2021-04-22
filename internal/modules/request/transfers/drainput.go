package transfers

import "github.com/Confialink/wallet-accounts/internal/modules/account/model"

type DRAInput interface {
	RevenueAccount() (*model.RevenueAccountModel, error)
}

type draInput struct {
	revenueAccount *model.RevenueAccountModel
}

func NewDraInput(revenueAccount *model.RevenueAccountModel) DRAInput {
	return &draInput{revenueAccount: revenueAccount}
}

func (d *draInput) RevenueAccount() (*model.RevenueAccountModel, error) {
	return d.revenueAccount, nil
}
