package form

type ConvertPreview struct {
	AccountIdFrom  *uint64 `json:"accountIdFrom" binding:"required"`
	AccountIdTo    *uint64 `json:"accountIdTo" binding:"required"`
	OutgoingAmount *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
	Note           *string `json:"note" binding:"omitempty,max=65535"`
}

type Convert struct {
	AccountIdFrom  *uint64 `json:"accountIdFrom" binding:"required"`
	AccountIdTo    *uint64 `json:"accountIdTo" binding:"required"`
	OutgoingAmount *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
	IncomingAmount *string `json:"incomingAmount" binding:"required,decimalGT=0"`
	Note           *string `json:"note" binding:"omitempty,max=65535"`
}
