package csv

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
)

type dataProcessor struct {
	request *model.Request
}

func (p *dataProcessor) getSourceAccountNumber() string {
	number, _ := p.request.SourceAccountNumber()
	return number
}

func (p *dataProcessor) getSourceAccountCurrency() string {
	if p.request.BaseCurrencyCode == nil {
		return ""
	}
	return *p.request.BaseCurrencyCode
}

func (p *dataProcessor) owtFeeType() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.Fee != nil {
		return *p.request.DataOwt.Fee.Name
	}
	return ""
}

func (p *dataProcessor) beneficiaryBankSwiftCode() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BankDetails != nil {
		return p.request.DataOwt.BankDetails.SwiftCode
	}
	return ""
}

func (p *dataProcessor) beneficiaryBankName() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BankDetails != nil {
		return p.request.DataOwt.BankDetails.BankName
	}
	return ""
}

func (p *dataProcessor) beneficiaryBankAddress() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BankDetails != nil {
		return p.request.DataOwt.BankDetails.Address
	}
	return ""
}

func (p *dataProcessor) beneficiaryBankLocation() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BankDetails != nil {
		return p.request.DataOwt.BankDetails.Location
	}
	return ""
}

func (p *dataProcessor) beneficiaryBankCountry() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BankDetails != nil {
		return *p.request.DataOwt.BankDetails.Country.Code
	}
	return ""
}

func (p *dataProcessor) beneficiaryBankAbaNumber() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BankDetails != nil {
		return p.request.DataOwt.BankDetails.AbaNumber
	}
	return ""
}

func (p *dataProcessor) beneficiaryName() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BeneficiaryCustomer != nil {
		return p.request.DataOwt.BeneficiaryCustomer.AccountName
	}
	return ""
}

func (p *dataProcessor) beneficiaryAddress() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BeneficiaryCustomer != nil {
		return p.request.DataOwt.BeneficiaryCustomer.Address
	}
	return ""
}

func (p *dataProcessor) beneficiaryIban() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.BeneficiaryCustomer != nil {
		return p.request.DataOwt.BeneficiaryCustomer.Iban
	}
	return ""
}

func (p *dataProcessor) refMessage() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.RefMessage != nil {
		return *p.request.DataOwt.RefMessage
	}
	return ""
}

func (p *dataProcessor) intermediaryBankSwift() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return p.request.DataOwt.IntermediaryBankDetails.SwiftCode
	}
	return ""
}

func (p *dataProcessor) intermediaryBankName() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return p.request.DataOwt.IntermediaryBankDetails.BankName
	}
	return ""
}

func (p *dataProcessor) intermediaryBankAddress() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return p.request.DataOwt.IntermediaryBankDetails.Address
	}
	return ""
}

func (p *dataProcessor) intermediaryBankLocation() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return p.request.DataOwt.IntermediaryBankDetails.Location
	}
	return ""
}

func (p *dataProcessor) intermediaryBankCountry() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return *p.request.DataOwt.IntermediaryBankDetails.Country.Code
	}
	return ""
}

func (p *dataProcessor) intermediaryBankAba() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return p.request.DataOwt.IntermediaryBankDetails.AbaNumber
	}
	return ""
}

func (p *dataProcessor) intermediaryBankIban() string {
	if *p.request.Subject == constants.SubjectTransferOutgoingWireTransfer && p.request.DataOwt.IntermediaryBankDetails != nil {
		return p.request.DataOwt.IntermediaryBankDetails.Iban
	}
	return ""
}
