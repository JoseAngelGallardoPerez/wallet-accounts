package model

import (
	"time"

	payoutMethodModel "github.com/Confialink/wallet-accounts/internal/modules/payment-method/model"
	paymentPeriodModel "github.com/Confialink/wallet-accounts/internal/modules/payment-period/model"
	"github.com/shopspring/decimal"
)

// TableName sets AccountType's table name to be `account_types`
func (*AccountType) TableName() string {
	return "account_types"
}

type AccountType struct {
	AccountTypePublic
	AccountTypePrivate
}

type AccountTypePublic struct {
	Name                      string                                 `json:"name" binding:"omitempty,alphanum"`
	CurrencyCode              string                                 `json:"currencyCode" binding:"activeCurrencyCode"`
	Currency                  *AccountTypeCurrency                   `json:"currency"`
	BalanceFeeAmount          *decimal.Decimal                       `json:"balanceFeeAmount"`
	BalanceChargeDay          *uint64                                `json:"balanceChargeDay"`
	BalanceLimitAmount        *decimal.Decimal                       `json:"balanceLimitAmount"`
	CreditLimitAmount         *decimal.Decimal                       `json:"creditLimitAmount"`
	CreditAnnualInterestRate  *decimal.Decimal                       `json:"creditAnnualInterestRate"`
	CreditPayoutMethod        *payoutMethodModel.PayoutMethodModel   `gorm:"save_associations:false;foreignkey:CreditPayoutMethodID;association_foreignkey:ID" json:"creditPayoutMethod" binding:"omitempty"`
	CreditPayoutMethodID      *uint64                                `json:"creditPayoutMethodId"`
	CreditChargePeriod        *paymentPeriodModel.PaymentPeriodModel `gorm:"save_associations:false;foreignkey:CreditChargePeriodID;association_foreignkey:ID" json:"creditChargePeriod" binding:"omitempty"`
	CreditChargePeriodID      *uint64                                `json:"creditChargePeriodId"`
	CreditChargeDay           *uint64                                `json:"creditChargeDay" binding:"omitempty,numeric,gte=1,lte=31"`
	CreditChargeMonth         *uint64                                `json:"creditChargeMonth" binding:"omitempty,numeric,gte=1,lte=12"`
	DepositAnnualInterestRate *decimal.Decimal                       `json:"depositAnnualInterestRate"`
	DepositPayoutMethod       *payoutMethodModel.PayoutMethodModel   `gorm:"save_associations:false;foreignkey:DepositPayoutMethodID;association_foreignkey:ID" json:"depositPayoutMethod" binding:"omitempty"`
	DepositPayoutMethodID     *uint64                                `json:"depositPayoutMethodId"`
	DepositPayoutPeriod       *paymentPeriodModel.PaymentPeriodModel `gorm:"save_associations:false;foreignkey:DepositPayoutPeriodID;association_foreignkey:ID" json:"depositPaymentPeriod" binding:"omitempty"`
	DepositPayoutPeriodID     *uint64                                `json:"depositPayoutPeriodId"`
	DepositPayoutDay          *uint64                                `json:"depositPayoutDay" binding:"omitempty,numeric,gte=1,lte=31"`
	DepositPayoutMonth        *uint64                                `json:"depositPayoutMonth" binding:"omitempty,numeric,gte=1,lte=12"`
	AutoNumberGeneration      *bool                                  `json:"autoNumberGeneration"`
	NumberPrefix              *string                                `json:"numberPrefix" binding:"omitempty"`
	MonthlyMaintenanceFee     *decimal.Decimal                       `json:"monthlyMaintenanceFee"`
}

type AccountTypePrivate struct {
	ID        uint64    `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AccountTypeEditable struct {
	Name                      string           `json:"name" binding:"omitempty,alphanum"`
	CurrencyCode              string           `json:"currencyCode" binding:"omitempty,activeCurrencyCode"`
	BalanceFeeAmount          *decimal.Decimal `json:"balanceFeeAmount" binding:"omitempty,numeric"`
	BalanceChargeDay          *uint64          `json:"balanceChargeDay" binding:"omitempty,numeric,gte=1,lte=31"`
	BalanceLimitAmount        *decimal.Decimal `json:"balanceLimitAmount" binding:"omitempty,numeric"`
	CreditLimitAmount         *decimal.Decimal `json:"creditLimitAmount" binding:"omitempty,numeric"`
	CreditAnnualInterestRate  *decimal.Decimal `json:"creditAnnualInterestRate" binding:"omitempty,numeric,gte=0,lte=100"`
	CreditPayoutMethodID      *uint64          `json:"creditPayoutMethodId" binding:"omitempty,numeric"`
	CreditChargePeriodID      *uint64          `json:"creditChargePeriodId" binding:"omitempty,numeric"`
	CreditChargeDay           *uint64          `json:"creditChargeDay" binding:"omitempty,numeric,gte=1,lte=31"`
	CreditChargeMonth         *uint64          `json:"creditChargeMonth" binding:"omitempty,numeric,gte=1,lte=12"`
	DepositAnnualInterestRate *decimal.Decimal `json:"depositAnnualInterestRate" binding:"omitempty,numeric,gte=0,lte=100"`
	DepositPayoutMethodID     *uint64          `json:"depositPayoutMethodId" binding:"omitempty,numeric"`
	DepositPayoutPeriodID     *uint64          `json:"depositPayoutPeriodId" binding:"omitempty,numeric"`
	DepositPayoutDay          *uint64          `json:"depositPayoutDay" binding:"omitempty,numeric,gte=1,lte=31"`
	DepositPayoutMonth        *uint64          `json:"depositPayoutMonth" binding:"omitempty,numeric,gte=1,lte=12"`
	AutoNumberGeneration      *bool            `json:"autoNumberGeneration"`
	NumberPrefix              *string          `json:"numberPrefix" binding:"omitempty"`
	MonthlyMaintenanceFee     *decimal.Decimal `json:"monthlyMaintenanceFee"`
	MinimumBalance            *decimal.Decimal `json:"minimumBalance"`
}

type AccountTypeCurrency struct {
	Id   *uint32 `json:"id"`
	Code *string `json:"code"`
}
