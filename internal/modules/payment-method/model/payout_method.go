package model

import "time"

// TableName sets PayoutMethod's table name to be `payout_methods`
func (PayoutMethodModel) TableName() string {
	return "payout_methods"
}

type PayoutMethodModel struct {
	PayoutMethodPublic
	PayoutMethodPrivate
}

type PayoutMethodPublic struct {
	Name string `gorm:"column:name" form:"name" json:"name"`
}

type PayoutMethodPrivate struct {
	ID        uint      `gorm:"primary_key" form:"id" json:"id"`
	Method    string    `json:"method"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
