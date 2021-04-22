package model

import (
	"time"

	countryModel "github.com/Confialink/wallet-accounts/internal/modules/country/model"
)

// TableName sets BankDetails's table name to be `bank_details`
func (BankDetailsModel) TableName() string {
	return "bank_details"
}

// BankDetailsModel bank details model
type BankDetailsModel struct {
	BankDetailsPublic
	BankDetailsPrivate
}

// BankDetailsPublic is used to create/edit a model
type BankDetailsPublic struct {
	SwiftCode string               `json:"swiftCode" binding:"required"`
	BankName  string               `json:"bankName" binding:"required"`
	Address   string               `json:"address" binding:"required"`
	Location  string               `json:"location" binding:"required"`
	Country   countryModel.Country `gorm:"foreignkey:CountryId;association_foreignkey:ID" json:"country"`
	CountryId uint64               `json:"countryId" binding:"required"`
	AbaNumber string               `json:"abaNumber"`
	Iban      string               `json:"iban"`
}

// BankDetailsPrivate contains fields assigned automatically
type BankDetailsPrivate struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
