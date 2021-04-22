package form

type DA struct {
	AccountId              uint64 `form:"accountId" json:"accountId" binding:"required"`
	Amount                 string `form:"amount" json:"amount" binding:"required"`
	Description            string `form:"description" json:"description" binding:"required"`
	CreditToRevenueAccount *bool  `form:"creditToRevenueAccount" json:"creditToRevenueAccount"`
}

type DAPreview struct {
	AccountId              uint64 `json:"accountId" binding:"required"`
	Amount                 string `json:"amount" binding:"required"`
	Description            string `json:"description" binding:"required"`
	CreditToRevenueAccount *bool  `json:"creditToRevenueAccount"`
}

func (f *DA) ToDAPreview() *DAPreview {
	return &DAPreview{
		AccountId:              f.AccountId,
		Amount:                 f.Amount,
		Description:            f.Description,
		CreditToRevenueAccount: f.CreditToRevenueAccount,
	}
}
