package form

type CFTPreview struct {
	AccountIdFrom  *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	CardIdTo       *uint32 `form:"cardIdTo" json:"cardIdTo" binding:"required"`
	OutgoingAmount *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
}

type CFT struct {
	*BaseTemplate
	AccountIdFrom  *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	CardIdTo       *uint32 `form:"cardIdTo" json:"cardIdTo" binding:"required"`
	OutgoingAmount *string `json:"outgoingAmount,omitempty" binding:"required,decimalGT=0"`
	Description    *string `json:"description,omitempty" binding:"required,max=65535"`
	IncomingAmount *string `json:"incomingAmount,omitempty" binding:"required,decimalGT=0"`
}

func (c *CFT) ToCFTPreview() *CFTPreview {
	return &CFTPreview{
		AccountIdFrom:  c.AccountIdFrom,
		CardIdTo:       c.CardIdTo,
		OutgoingAmount: c.OutgoingAmount,
	}
}

func (c CFT) TemplateData() interface{} {
	c.BaseTemplate = nil
	c.IncomingAmount = nil
	return c
}
