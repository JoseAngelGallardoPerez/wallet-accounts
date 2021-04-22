package model

import (
	"time"
)

// TableName sets IwtBankAccount's table name to be `iwt_bank_accounts`
func (IwtBankAccountModel) TableName() string {
	return "iwt_bank_accounts"
}

// IwtBankAccountModel iwt bank account model
type IwtBankAccountModel struct {
	IwtBankAccountPublic
	IwtBankAccountPrivate
}

// IwtBankAccountPublic is used to create/edit a model
type IwtBankAccountPublic struct {
	CurrencyCode              string                    `json:"currencyCode" binding:"required"`
	IsIwtEnabled              *bool                     `json:"isIwtEnabled" binding:"required"`
	BeneficiaryBankDetails    *BankDetailsModel         `gorm:"foreignkey:BeneficiaryBankDetailsId;association_foreignkey:ID" json:"beneficiaryBankDetails"`
	BeneficiaryBankDetailsId  uint64                    `json:"beneficiaryBankDetailsId"`
	BeneficiaryCustomer       *BeneficiaryCustomerModel `gorm:"foreignkey:BeneficiaryCustomerId;association_foreignkey:ID" json:"beneficiaryCustomer"`
	BeneficiaryCustomerId     uint64                    `json:"beneficiaryCustomerId"`
	IntermediaryBankDetails   *BankDetailsModel         `gorm:"foreignkey:IntermediaryBankDetailsId;association_foreignkey:ID" json:"intermediaryBankDetails"`
	IntermediaryBankDetailsId *uint64                   `json:"intermediaryBankDetailsId"`
	AdditionalInstructions    *string                   `json:"additionalInstructions"`
}

// IwtBankAccountPrivate contains fields assigned automatically
type IwtBankAccountPrivate struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
