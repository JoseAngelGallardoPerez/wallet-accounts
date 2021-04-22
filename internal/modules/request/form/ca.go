package form

type CA struct {
	AccountId               uint64 `form:"accountId" json:"accountId" binding:"required"`
	Amount                  string `form:"amount" json:"amount" binding:"required"`
	Description             string `form:"description" json:"description" binding:"required"`
	DebitFromRevenueAccount *bool  `form:"debitFromRevenueAccount" json:"debitFromRevenueAccount"`
	ApplyIwtFee             *bool  `form:"applyIwtFee" json:"applyIwtFee"`
}

type CAPreview struct {
	AccountId               uint64 `json:"accountId" binding:"required"`
	Amount                  string `json:"amount" binding:"required"`
	Description             string `json:"description" binding:"required"`
	DebitFromRevenueAccount *bool  `json:"debitFromRevenueAccount"`
	ApplyIwtFee             *bool  `json:"applyIwtFee"`
}

func (f *CA) ToCAPreview() *CAPreview {
	return &CAPreview{
		AccountId:               f.AccountId,
		Amount:                  f.Amount,
		Description:             f.Description,
		DebitFromRevenueAccount: f.DebitFromRevenueAccount,
		ApplyIwtFee:             f.ApplyIwtFee,
	}
}
