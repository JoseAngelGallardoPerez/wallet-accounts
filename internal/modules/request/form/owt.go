package form

import (
	"reflect"

	bankDetailsModel "github.com/Confialink/wallet-accounts/internal/modules/bank-details/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/go-playground/validator/v10"
)

type OWTPreview struct {
	AccountIdFrom         *uint64 `form:"accountIdFrom" json:"accountIdFrom" binding:"required"`
	ReferenceCurrencyCode *string `form:"referenceCurrencyCode" json:"referenceCurrencyCode" binding:"required"`
	OutgoingAmount        *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
	FeeId                 *uint64 `json:"feeId"`
}

type OWT struct {
	*BaseTemplate
	AccountIdFrom              *uint64 `json:"accountIdFrom" binding:"required"`
	ReferenceCurrencyCode      *string `json:"referenceCurrencyCode" binding:"required"`
	OutgoingAmount             *string `json:"outgoingAmount" binding:"required,decimalGT=0"`
	ConfirmTotalOutgoingAmount *string `json:"confirmTotalOutgoingAmount,omitempty" binding:"required,decimal"`
	Description                *string `json:"description"`
	RefMessage                 *string `json:"refMessage" binding:"required"`
	BankSwiftBic               *string `json:"bankSwiftBic" binding:"required"`
	BankName                   *string `json:"bankName" binding:"required"`
	BankAddress                *string `json:"bankAddress" binding:"required"`
	BankLocation               *string `json:"bankLocation" binding:"required"`
	BankCountryId              *uint64 `json:"bankCountryId" binding:"required"`
	BankAbaRtn                 *string `json:"bankAbaRtn" binding:"notRequiredAlphanumericPointer"`
	CustomerName               *string `json:"customerName" binding:"required"`
	CustomerAddress            *string `json:"customerAddress" binding:"required"`
	CustomerAccIban            *string `json:"customerAccIban" binding:"required"`
	IsIntermediaryBankRequired *bool   `json:"isIntermediaryBankRequired"`
	FeeId                      *uint64 `json:"feeId"`

	IntermediaryBankSwiftBic  *string `json:"intermediaryBankSwiftBic"`
	IntermediaryBankName      *string `json:"intermediaryBankName"`
	IntermediaryBankAddress   *string `json:"intermediaryBankAddress"`
	IntermediaryBankLocation  *string `json:"intermediaryBankLocation"`
	IntermediaryBankCountryId *uint64 `json:"intermediaryBankCountryId"`
	IntermediaryBankAbaRtn    *string `json:"intermediaryBankAbaRtn" binding:"notRequiredAlphanumericPointer"`
	IntermediaryBankAccIban   *string `json:"intermediaryBankAccIban"`
}

func (o *OWT) ToOWTPreview() *OWTPreview {
	return &OWTPreview{
		AccountIdFrom:         o.AccountIdFrom,
		ReferenceCurrencyCode: o.ReferenceCurrencyCode,
		OutgoingAmount:        o.OutgoingAmount,
		FeeId:                 o.FeeId,
	}
}

func (o OWT) TemplateData() interface{} {
	o.BaseTemplate = nil
	o.ConfirmTotalOutgoingAmount = nil
	return o
}

func (o *OWT) NewDataOwt() *model.DataOwt {
	data := &model.DataOwt{
		SourceAccountId:         o.AccountIdFrom,
		DestinationCurrencyCode: o.ReferenceCurrencyCode,
		RefMessage:              o.RefMessage,
		FeeId:                   o.FeeId,
	}

	bankDetails := &bankDetailsModel.BankDetailsModel{
		BankDetailsPublic: bankDetailsModel.BankDetailsPublic{
			BankName:  *o.BankName,
			Location:  *o.BankLocation,
			Address:   *o.BankAddress,
			CountryId: *o.BankCountryId,
			AbaNumber: *o.BankAbaRtn,
			SwiftCode: *o.BankSwiftBic,
		},
	}
	data.BankDetails = bankDetails

	customerDetails := &bankDetailsModel.BeneficiaryCustomerModel{
		BeneficiaryCustomerPublic: bankDetailsModel.BeneficiaryCustomerPublic{
			Address:     *o.CustomerAddress,
			Iban:        *o.CustomerAccIban,
			AccountName: *o.CustomerName,
		},
	}
	data.BeneficiaryCustomer = customerDetails
	if o.IsIntermediaryBankRequired != nil && *o.IsIntermediaryBankRequired {
		intermediaryBankDetails := &bankDetailsModel.BankDetailsModel{
			BankDetailsPublic: bankDetailsModel.BankDetailsPublic{
				BankName:  *o.IntermediaryBankName,
				SwiftCode: *o.IntermediaryBankSwiftBic,
				Iban:      *o.IntermediaryBankAccIban,
				Address:   *o.IntermediaryBankAddress,
				Location:  *o.IntermediaryBankLocation,
				AbaNumber: *o.IntermediaryBankAbaRtn,
				CountryId: *o.IntermediaryBankCountryId,
			},
		}
		data.IntermediaryBankDetails = intermediaryBankDetails
	}

	return data
}

func OwtStructLevelValidation(sl validator.StructLevel) {
	owt := sl.Current().Interface().(OWT)
	if owt.IsIntermediaryBankRequired != nil && *owt.IsIntermediaryBankRequired {
		tag := "required"
		if owt.IntermediaryBankSwiftBic == nil || *owt.IntermediaryBankSwiftBic == "" {
			sl.ReportError(reflect.ValueOf(owt.IntermediaryBankName), "IntermediaryBankSwiftBic", "intermediaryBankSwiftBic", tag, "")
		}
		if owt.IntermediaryBankName == nil || *owt.IntermediaryBankName == "" {
			sl.ReportError(reflect.ValueOf(owt.IntermediaryBankName), "IntermediaryBankName", "intermediaryBankName", tag, "")
		}
		if owt.IntermediaryBankAddress == nil || *owt.IntermediaryBankAddress == "" {
			sl.ReportError(reflect.ValueOf(owt.IntermediaryBankAddress), "IntermediaryBankAddress", "intermediaryBankAddress", tag, "")
		}
		if owt.IntermediaryBankLocation == nil || *owt.IntermediaryBankLocation == "" {
			sl.ReportError(reflect.ValueOf(owt.IntermediaryBankLocation), "IntermediaryBankLocation", "intermediaryBankLocation", tag, "")
		}
		if owt.IntermediaryBankCountryId == nil || *owt.IntermediaryBankCountryId == 0 {
			sl.ReportError(reflect.ValueOf(owt.IntermediaryBankCountryId), "IntermediaryBankCountryId", "intermediaryBankCountryId", tag, "")
		}
		if owt.IntermediaryBankAccIban == nil || *owt.IntermediaryBankAccIban == "" {
			sl.ReportError(reflect.ValueOf(owt.IntermediaryBankAccIban), "IntermediaryBankAccIban", "intermediaryBankAccIban", tag, "")
		}
	}
}
