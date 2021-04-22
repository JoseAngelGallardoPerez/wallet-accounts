package form

import "github.com/Confialink/wallet-accounts/internal/modules/request/constants"

type TransferFee struct {
	Name           *string                   `json:"name" binding:"required"`
	UserGroupIds   []uint64                  `json:"userGroupIds" binding:"required"`
	RequestSubject *constants.Subject        `json:"requestSubject" binding:"required"`
	Parameters     TransferFeeParametersList `json:"parameters" binding:"dive"`
}

type UpdateTransferFee struct {
	Name       *string                   `json:"name" binding:"omitempty"`
	Parameters TransferFeeParametersList `json:"parameters" binding:"dive"`
	Relations  []*UserGroupRelation      `json:"relations" binding:"dive"`
}

type UserGroupRelation struct {
	UserGroupId *uint64 `json:"userGroupId" binding:"required"`
	//indicates relation state between user group and transfer fee
	//i.e. should the user group be attached or detached
	Attached *bool `json:"attached" binding:"exists"`
}
