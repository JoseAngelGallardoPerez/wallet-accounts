package model

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	bankDetailsModel "github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	transferFeeModel "github.com/Confialink/wallet-accounts/internal/modules/fee/model"
)

type DataOwt struct {
	Id                        *uint64                                    `json:"_,"`
	RequestId                 *uint64                                    `json:"requestId"`
	SourceAccountId           *uint64                                    `json:"sourceAccountId"`
	SourceAccount             *accountModel.Account                      `json:"_"`
	DestinationCurrencyCode   *string                                    `json:"destinationCurrencyCode"`
	BankDetailsId             *uint64                                    `json:"_"`
	BeneficiaryCustomerId     *uint64                                    `json:"_"`
	IntermediaryBankDetailsId *uint64                                    `json:"_"`
	RefMessage                *string                                    `json:"refMessage"`
	BankDetails               *bankDetailsModel.BankDetailsModel         `gorm:"foreignkey:BankDetailsId;association_foreignkey:ID" json:"bankDetails"`
	BeneficiaryCustomer       *bankDetailsModel.BeneficiaryCustomerModel `gorm:"foreignkey:BeneficiaryCustomerId;association_foreignkey:ID" json:"beneficiaryCustomer"`
	IntermediaryBankDetails   *bankDetailsModel.BankDetailsModel         `gorm:"foreignkey:IntermediaryBankDetailsId;association_foreignkey:ID" json:"intermediaryBankDetails"`
	FeeId                     *uint64                                    `json:"-"`
	Fee                       *transferFeeModel.TransferFee              `gorm:"foreignkey:FeeId;association_foreignkey:Id" json:"_"`
}

func (*DataOwt) TableName() string {
	return "request_data_owt"
}

func (d *DataOwt) GetSourceAccountId() uint64 {
	return *d.SourceAccountId
}
