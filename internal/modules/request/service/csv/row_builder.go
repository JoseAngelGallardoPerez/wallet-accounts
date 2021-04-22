package csv

import (
	"strconv"
	"strings"

	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/pkg/errors"

	"github.com/Confialink/wallet-pkg-utils/timefmt"

	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/syssettings"
)

const (
	profileTypePersonal  = "Personal"
	profileTypeCorporate = "Corporate"
)

type RowBuilder struct {
	request       *model.Request
	timeSettings  *syssettings.TimeSettings
	dataProcessor *dataProcessor
	roleName      string
}

func NewRowBuilder(request *model.Request, timeSettings *syssettings.TimeSettings, roleName string) *RowBuilder {
	return &RowBuilder{
		request,
		timeSettings,
		&dataProcessor{request},
		roleName,
	}
}

func (b *RowBuilder) Call() ([]string, error) {
	switch b.roleName {
	case "admin", "root":
		return b.getAdminRow()
	default:
		return nil, errors.New("undefined role name " + b.roleName)
	}
}

func (b *RowBuilder) getAdminRow() ([]string, error) {
	return []string{
		b.id(),
		b.createdAt(),
		b.statusChangedAt(),
		b.userName(),
		b.profileType(),
		b.companyName(),
		b.firstName(),
		b.lastName(),
		//b.country(),
		//b.zipCode(),
		//b.state(),
		//b.city(),
		//b.address(),
		//b.address2ndLine(),
		b.userGroup(),
		b.payFromAccountNumber(),
		b.accountFromCurrency(),
		b.description(),
		b.subject(),
		b.status(),
		b.owtFeeType(),
		b.paymentAmount(),
		b.paymentCurrency(),
		b.beneficiaryBankSwiftCode(),
		b.beneficiaryBankName(),
		b.beneficiaryBankAddress(),
		b.beneficiaryBankLocation(),
		b.beneficiaryBankCountry(),
		b.beneficiaryBankAbaNumber(),
		b.beneficiaryName(),
		b.beneficiaryAddress(),
		b.beneficiaryIban(),
		b.refMessage(),
		b.intermediaryBankSwift(),
		b.intermediaryBankName(),
		b.intermediaryBankAddress(),
		b.intermediaryBankLocation(),
		b.intermediaryBankCountry(),
		b.intermediaryBankAba(),
		b.intermediaryBankIban(),
	}, nil
}

func (b *RowBuilder) id() string {
	return strconv.FormatUint(*b.request.Id, 10)
}

func (b *RowBuilder) statusChangedAt() string {
	if b.request.StatusChangedAt == nil {
		return ""
	}
	return timefmt.Format(*b.request.StatusChangedAt, b.timeSettings.DateTimeFormat, b.timeSettings.Timezone)
}

func (b *RowBuilder) createdAt() string {
	return timefmt.Format(*b.request.CreatedAt, b.timeSettings.DateTimeFormat, b.timeSettings.Timezone)
}

func (b *RowBuilder) updatedAt() string {
	if *b.request.Status == constants.StatusPending || *b.request.Status == constants.StatusNew {
		return ""
	}
	return timefmt.Format(*b.request.UpdatedAt, b.timeSettings.DateTimeFormat, b.timeSettings.Timezone)
}

func (b *RowBuilder) userName() string {
	return b.request.FullUser.Username
}

func (b *RowBuilder) profileType() string {
	if b.request.FullUser.IsCorporate {
		return profileTypeCorporate
	}
	return profileTypePersonal
}

func (b *RowBuilder) companyName() string {
	return b.request.FullUser.Company.CompanyName
}

func (b *RowBuilder) firstName() string {
	return b.request.FullUser.FirstName
}

func (b *RowBuilder) lastName() string {
	return b.request.FullUser.LastName
}

func (b *RowBuilder) country() string {
	//return b.request.FullUser.PhysicalAddress.PaCountryIso2
	return "n/a"
}

func (b *RowBuilder) zipCode() string {
	//return b.request.FullUser.PhysicalAddress.PaZipPostalCode
	return "n/a"
}

func (b *RowBuilder) state() string {
	//return b.request.FullUser.PhysicalAddress.PaStateProvRegion
	return "n/a"
}

func (b *RowBuilder) city() string {
	//return b.request.FullUser.PhysicalAddress.PaCity
	return "n/a"
}

func (b *RowBuilder) address() string {
	//return b.request.FullUser.PhysicalAddress.PaAddress
	return "n/a"
}

func (b *RowBuilder) address2ndLine() string {
	//return b.request.FullUser.PhysicalAddress.PaAddress2NdLine
	return "n/a"
}

func (b *RowBuilder) userGroup() string {
	if b.request.FullUser.Group != nil {
		return b.request.FullUser.Group.Name
	}
	return ""
}

func (b *RowBuilder) payFromAccountNumber() string {
	return b.dataProcessor.getSourceAccountNumber()
}

func (b *RowBuilder) accountFromCurrency() string {
	return b.dataProcessor.getSourceAccountCurrency()
}

func (b *RowBuilder) description() string {
	if b.request.Description != nil {
		return *b.request.Description
	}
	return ""
}

func (b *RowBuilder) subject() string {
	return strings.ToLower(string(*b.request.Subject))
}

func (b *RowBuilder) status() string {
	return *b.request.Status
}

func (b *RowBuilder) owtFeeType() string {
	return b.dataProcessor.owtFeeType()
}

func (b *RowBuilder) paymentAmount() string {
	return b.request.Amount.String()
}

func (b *RowBuilder) paymentCurrency() string {
	return *b.request.ReferenceCurrencyCode
}

func (b *RowBuilder) beneficiaryBankSwiftCode() string {
	return b.dataProcessor.beneficiaryBankSwiftCode()
}

func (b *RowBuilder) beneficiaryBankName() string {
	return b.dataProcessor.beneficiaryBankName()
}

func (b *RowBuilder) beneficiaryBankAddress() string {
	return b.dataProcessor.beneficiaryBankAddress()
}

func (b *RowBuilder) beneficiaryBankLocation() string {
	return b.dataProcessor.beneficiaryBankLocation()
}

func (b *RowBuilder) beneficiaryBankCountry() string {
	return b.dataProcessor.beneficiaryBankCountry()
}

func (b *RowBuilder) beneficiaryBankAbaNumber() string {
	return b.dataProcessor.beneficiaryBankAbaNumber()
}

func (b *RowBuilder) beneficiaryName() string {
	return b.dataProcessor.beneficiaryName()
}

func (b *RowBuilder) beneficiaryAddress() string {
	return b.dataProcessor.beneficiaryAddress()
}

func (b *RowBuilder) beneficiaryIban() string {
	return b.dataProcessor.beneficiaryIban()
}

func (b *RowBuilder) refMessage() string {
	return b.dataProcessor.refMessage()
}

func (b *RowBuilder) intermediaryBankSwift() string {
	return b.dataProcessor.intermediaryBankSwift()
}

func (b *RowBuilder) intermediaryBankName() string {
	return b.dataProcessor.intermediaryBankName()
}

func (b *RowBuilder) intermediaryBankAddress() string {
	return b.dataProcessor.intermediaryBankAddress()
}

func (b *RowBuilder) intermediaryBankLocation() string {
	return b.dataProcessor.intermediaryBankLocation()
}

func (b *RowBuilder) intermediaryBankCountry() string {
	return b.dataProcessor.intermediaryBankCountry()
}

func (b *RowBuilder) intermediaryBankAba() string {
	return b.dataProcessor.intermediaryBankAba()
}

func (b *RowBuilder) intermediaryBankIban() string {
	return b.dataProcessor.intermediaryBankIban()
}

func (b *RowBuilder) availableAmountSnapshot() string {
	var amount string
	for _, s := range b.request.BalanceSnapshots {
		if s.BalanceType.Name == "account" {
			val, err := s.GetValue()
			if err != nil {
				return amount
			}
			amount = val.AvailableAmount.String()
		}
	}
	return amount
}

func (b *RowBuilder) balanceDifference() string {
	var amount string
	if len(b.request.BalanceDifference) > 0 {
		amount = b.request.BalanceDifference[0].Difference.String()
	}
	return amount
}
