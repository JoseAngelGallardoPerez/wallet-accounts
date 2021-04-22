package form

type MoneyRequestTBUPreview struct {
	AccountIdFrom  *uint64 `json:"accountIdFrom" binding:"required"`
	MoneyRequestId *uint64 `json:"moneyRequestId" binding:"required"`
}
