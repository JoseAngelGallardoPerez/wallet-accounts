package tan

type Count struct {
	UserId   string `json:"userId"`
	Quantity uint   `json:"quantity"`
}

type CreateForm struct {
	Quantity           uint   `json:"quantity" form:"quantity" binding:"required"`
	CancelOld          bool   `json:"cancelOld" form:"cancelOld"`
	NotificationMethod string `json:"notificationMethod"`
}
