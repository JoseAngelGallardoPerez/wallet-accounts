package types

import (
	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	cardModel "github.com/Confialink/wallet-accounts/internal/modules/card/model"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
)

var zero = decimal.NewFromInt(0)

type Detail struct {
	Purpose          constants.Purpose `json:"purpose"`
	Amount           decimal.Decimal   `json:"amount"`
	CurrencyCode     string            `json:"currencyCode"`
	Transaction      *transactionModel.Transaction
	AccountId        *uint64
	RevenueAccountId *uint64
	CardId           *uint32

	Account        *accountModel.Account
	RevenueAccount *accountModel.RevenueAccountModel
	Card           *cardModel.Card
}

func (d *Detail) GoString() string {
	return fmt.Sprintf("Detail{Purpose: \"%s\" Amount: %s, GetCurrencyCode: \"%s\"}", d.Purpose, d.Amount, d.CurrencyCode)
}

// IsCredit indicates whether detail amount is positive (i.e. credit)
func (d *Detail) IsCredit() bool {
	return d.Amount.GreaterThan(zero)
}

// IsDebit indicates whether detail amount is negative (i.e. debit)
func (d *Detail) IsDebit() bool {
	return d.Amount.LessThan(zero)
}

type Details map[constants.Purpose]*Detail

func (d Details) GoString() string {
	result := "Details {\n"
	for purpose, detail := range d {
		result += "\t " + purpose.String() + ": " + detail.GoString() + ",\n"
	}
	result += "\n}"
	return result
}

// ByPurposes retrieves certain details by purposes
func (d Details) ByPurposes(purposes ...constants.Purpose) Details {
	result := make(Details)
	for _, purpose := range purposes {
		if detail, ok := d[purpose]; ok {
			result[purpose] = detail
		}
	}
	return result
}

// ByPurpose retrieves certain detail by purpose
func (d Details) ByPurpose(purpose constants.Purpose) *Detail {
	if detail, ok := d[purpose]; ok {
		return detail
	}

	return nil
}

// ByAccountId retrieves certain details by account id
func (d Details) ByAccountId(accountId uint64) Details {
	result := make(Details)
	for _, detail := range d {
		if detail.AccountId != nil && *detail.AccountId == accountId {
			result[detail.Purpose] = detail
		}
	}
	return result
}

// SumByAccountId calculates total amount by account id
func (d Details) SumByAccountId(accountId uint64) decimal.Decimal {
	result := decimal.New(0, 0)
	for _, detail := range d {
		if detail.AccountId != nil && *detail.AccountId == accountId {
			result = result.Add(detail.Amount)
		}
	}
	return result
}

// TotalAccountDebit provides total debited amount by account id
func (d Details) TotalAccountDebit(accountId uint64) decimal.Decimal {
	return d.reduce(func(detail *Detail) decimal.Decimal {
		if detail.AccountId != nil && *detail.AccountId == accountId && detail.IsDebit() {
			return detail.Amount
		}
		return zero
	})
}

// TotalAccountDebit provides total debited amount by account id
func (d Details) TotalAccountCredit(accountId uint64) decimal.Decimal {
	return d.reduce(func(detail *Detail) decimal.Decimal {
		if detail.AccountId != nil && *detail.AccountId == accountId && detail.IsCredit() {
			return detail.Amount
		}
		return zero
	})
}

func (d Details) reduce(f func(*Detail) decimal.Decimal) decimal.Decimal {
	result := decimal.NewFromInt(0)
	for _, detail := range d {
		result = result.Add(f(detail))
	}
	return result
}
