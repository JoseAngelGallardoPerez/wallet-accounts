package form

type DRA struct {
	RevenueAccountId uint64  `form:"revenueAccountId" json:"revenueAccountId" binding:"required"`
	Amount           string  `form:"amount" json:"amount" binding:"required"`
	Description      *string `form:"description" json:"description,omitempty"`
}

type DRAPreview struct {
	RevenueAccountId uint64  `json:"revenueAccountId" binding:"required"`
	Amount           string  `json:"amount" binding:"required"`
	Description      *string `json:"description,omitempty"`
}

func (f *DRA) ToDRAPreview() *DRAPreview {
	return &DRAPreview{
		RevenueAccountId: f.RevenueAccountId,
		Amount:           f.Amount,
		Description:      f.Description,
	}
}
