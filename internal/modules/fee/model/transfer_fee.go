package model

import (
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-pkg-utils/value"
	"github.com/shopspring/decimal"
)

type TransferFee struct {
	Id             *uint64 `gorm:"primary_key" json:"id"`
	Name           *string
	RequestSubject *constants.Subject
	Relations      []*TransferFeeUserGroup `gorm:"foreignkey:TransferFeeId"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TransferFeeParameters struct {
	Id            *uint64          `gorm:"primary_key" json:"id"`
	TransferFeeId *uint64          `json:"transferFeeId"`
	CurrencyCode  *string          `json:"currencyCode"`
	Base          *decimal.Decimal `json:"base"`
	Min           *decimal.Decimal `json:"min"`
	Percent       *decimal.Decimal `json:"percent"`
	Max           *decimal.Decimal `json:"max"`
}

type TransferFeeUserGroup struct {
	TransferFeeId *uint64 `json:"-"`
	UserGroupId   *uint64 `json:"userGroupId"`
}

func (*TransferFeeUserGroup) TableName() string {
	return "transfer_fees_user_groups"
}

func (*TransferFeeParameters) TableName() string {
	return "transfer_fees_parameters"
}

func (t *TransferFee) MarshalJSON() ([]byte, error) {

	userGroups := make([]uint64, len(t.Relations))
	for i, relation := range t.Relations {
		userGroups[i] = value.FromUint64(relation.UserGroupId)
	}

	return json.Marshal(map[string]interface{}{
		"id":             t.Id,
		"name":           t.Name,
		"requestSubject": t.RequestSubject,
		"userGroups":     userGroups,
		"createdAt":      t.CreatedAt,
	})
}

func (t *TransferFeeParameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":            t.Id,
		"transferFeeId": t.TransferFeeId,
		"currencyCode":  t.CurrencyCode,
		"base":          t.Base,
		"min":           t.Min,
		"percent":       t.Percent,
		"max":           t.Max,
	})
}

func (*TransferFee) TableName() string {
	return "transfer_fees"
}
