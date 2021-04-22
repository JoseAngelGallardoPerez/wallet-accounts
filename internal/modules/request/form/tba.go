package form

type TBAPreview struct {
	AccountIdFrom  *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	AccountIdTo    *uint64 `form:"accountIdTo" json:"accountIdTo" binding:"required"`
	OutgoingAmount *string `json:"outgoingAmount" binding:"omitempty,decimalGT=0"`
	IncomingAmount *string `json:"incomingAmount" binding:"omitempty"`
}

type TBA struct {
	AccountIdFrom  *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	AccountIdTo    *uint64 `form:"accountIdTo" json:"accountIdTo" binding:"required"`
	OutgoingAmount *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
	Description    *string `json:"description" binding:"required,max=65535"`
	IncomingAmount *string `json:"incomingAmount" binding:"required,decimalGT=0"`
}

func (f *TBA) ToTBAPreview() *TBAPreview {
	return &TBAPreview{
		AccountIdTo:    f.AccountIdTo,
		AccountIdFrom:  f.AccountIdFrom,
		OutgoingAmount: f.OutgoingAmount,
	}
}
