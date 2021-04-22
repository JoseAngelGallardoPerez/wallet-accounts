package model

import "time"

// TableName sets PaymentPeriod's table name to be `payment_periods`
func (PaymentPeriodModel) TableName() string {
	return "payment_periods"
}

type PaymentPeriodModel struct {
	PaymentPeriodPublic
	PaymentPeriodPrivate
}

type PaymentPeriodPublic struct {
	Name string `gorm:"column:name" form:"name" json:"name"`
}

type PaymentPeriodPrivate struct {
	ID        uint      `gorm:"primary_key" form:"id" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
