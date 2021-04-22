package model

import (
	"time"
)

// TableName sets BeneficiaryCustomer's table name to be `beneficiary_customers`
func (BeneficiaryCustomerModel) TableName() string {
	return "beneficiary_customers"
}

// BeneficiaryCustomerModel beneficiary customer model
type BeneficiaryCustomerModel struct {
	BeneficiaryCustomerPublic
	BeneficiaryCustomerPrivate
}

// BeneficiaryCustomerPublic is used to create/edit a model
type BeneficiaryCustomerPublic struct {
	AccountName string `json:"accountName" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Iban        string `json:"iban"`
}

// BeneficiaryCustomerPrivate contains fields assigned automatically
type BeneficiaryCustomerPrivate struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
