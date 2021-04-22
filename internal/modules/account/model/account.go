package model

import (
	"github.com/pkg/errors"
	"time"

	accountTypeModel "github.com/Confialink/wallet-accounts/internal/modules/account-type/model"
	userModel "github.com/Confialink/wallet-accounts/internal/modules/user/model"
	"github.com/shopspring/decimal"
)

// TableName sets Account's table name to be `accounts`
func (*Account) TableName() string {
	return "accounts"
}

const AccountStatusIsActiveTrue = "1"

// Account account model
type Account struct {
	AccountPublic
	AccountPrivate
}

// AccountPublic is used to create a new account
type AccountPublic struct {
	Number            string                        `json:"number" binding:"omitempty,alphanum,max=28,accountNumberUnique"`
	Type              *accountTypeModel.AccountType `gorm:"foreignkey:TypeId;association_foreignkey:ID;association_autoupdate:false" json:"type"`
	TypeID            uint64                        `json:"typeId" binding:"required"`
	UserId            string                        `json:"userId" binding:"required"`
	Description       *string                       `json:"description"`
	IsActive          *bool                         `json:"isActive"`
	InitialBalance    *decimal.Decimal              `json:"initialBalance"`
	AllowWithdrawals  *bool                         `json:"allowWithdrawals"`
	AllowDeposits     *bool                         `json:"allowDeposits"`
	MaturityDate      *time.Time                    `json:"maturityDate"`
	PayoutDay         *uint64                       `json:"payoutDay" binding:"omitempty,gte=1,lte=31"`
	InterestAccountId *uint64                       `json:"interestAccountId" binding:"omitempty"`
	InterestAccount   *Account                      `gorm:"foreignkey:InterestAccountId;association_foreignkey:ID;association_autoupdate:false;association_save_reference:false" json:"interestAccount" binding:"omitempty"`
	User              *userModel.User               `json:"user"`
}

// AccountPrivate contains fields assigned automatically
type AccountPrivate struct {
	ID              uint64          `gorm:"primary_key" json:"id"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
	AvailableAmount decimal.Decimal `json:"availableAmount"`
	Balance         decimal.Decimal `json:"balance"`
}

// AccountEditable contains fields can be modified
type AccountEditable struct {
	Description       *string    `json:"description"`
	IsActive          *bool      `json:"isActive"`
	AllowWithdrawals  *bool      `json:"allowWithdrawals"`
	AllowDeposits     *bool      `json:"allowDeposits"`
	MaturityDate      *time.Time `json:"maturityDate"`
	PayoutDay         *uint64    `json:"payoutDay" binding:"omitempty,gte=1,lte=31"`
	InterestAccountId *uint64    `json:"interestAccountId" binding:"omitempty"`
}

func (a *Account) BeforeCreate() {
	if a.AvailableAmount.Equal(decimal.Zero) {
		a.AvailableAmount = a.Balance
	}
}

func (a *Account) CurrentBalance() (decimal.Decimal, error) {
	return a.Balance, nil
}

func (a *Account) AvailableBalance() (decimal.Decimal, error) {
	return a.AvailableAmount, nil
}

func (a *Account) GetCurrencyCode() (string, error) {
	if a.Type == nil {
		return "", errors.New("account type must be loaded to be able to access currency code")
	}
	return a.Type.CurrencyCode, nil
}

func (a *Account) TypeName() string {
	return "account"
}

func (a *Account) GetId() *uint64 {
	return &a.ID
}

func (a *Account) GetUserId() *string {
	return &a.UserId
}

// ValueByFieldName returns field value by its name
//func (a *Account) ValueByFieldName(name string) (interface{}, bool) {
//	switch name {
//	case "Number", "number":
//		return a.Number, true
//	case "Type", "type":
//		return a.Type, true
//	case "TypeID", "type_id":
//		return a.TypeID, true
//	case "UserId", "user_id":
//		return a.UserId, true
//	case "Description", "description":
//		return a.Description, true
//	case "IsActive", "is_active":
//		return a.IsActive, true
//	case "InitialBalance", "initial_balance":
//		return a.InitialBalance, true
//	case "AllowWithdrawals", "allow_withdrawals":
//		return a.AllowWithdrawals, true
//	case "AllowDeposits", "allow_deposits":
//		return a.AllowDeposits, true
//	case "MaturityDate", "maturity_date":
//		return a.MaturityDate, true
//	case "PayoutDay", "payout_day":
//		return a.PayoutDay, true
//	case "InterestAccountId", "interest_account_id":
//		return a.InterestAccountId, true
//	case "InterestAccount", "interest_account":
//		return a.InterestAccount, true
//	case "User", "user":
//		return a.User, true
//	case "ID", "id":
//		return a.ID, true
//	case "CreatedAt", "created_at":
//		return a.CreatedAt, true
//	case "UpdatedAt", "updated_at":
//		return a.UpdatedAt, true
//	case "AvailableAmount", "available_amount":
//		return a.AvailableAmount, true
//	case "CurrentBalance", "balance":
//		return a.CurrentBalance, true
//	}
//	return nil, false
//}
