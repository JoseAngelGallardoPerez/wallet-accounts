package form

type TBUPreview struct {
	AccountIdFrom   *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	AccountNumberTo *string `form:"accountNumberTo" json:"accountNumberTo" binding:"required"`
	OutgoingAmount  *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
}

type TBUReceive struct {
	AccountIdTo *uint64 `form:"accountIdTo" json:"accountIdTo" binding:"required"`
}

type TBU struct {
	*BaseTemplate
	AccountIdFrom   *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	AccountNumberTo *string `form:"accountNumberTo" json:"accountNumberTo" binding:"required"`
	OutgoingAmount  *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
	Description     *string `json:"description,omitempty" binding:"omitempty,max=65535"`
	IncomingAmount  *string `json:"incomingAmount,omitempty" binding:"required,decimalGT=0"`
}

func (t *TBU) ToTBUPreview() *TBUPreview {
	return &TBUPreview{
		AccountNumberTo: t.AccountNumberTo,
		AccountIdFrom:   t.AccountIdFrom,
		OutgoingAmount:  t.OutgoingAmount,
	}
}

func (t TBU) TemplateData() interface{} {
	t.BaseTemplate = nil
	t.IncomingAmount = nil
	return t
}
