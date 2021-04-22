package model

import (
	"github.com/Confialink/wallet-accounts/internal/conv"
	"github.com/Confialink/wallet-pkg-utils/value"
	"encoding/json"
	"time"

	"github.com/Confialink/wallet-accounts/internal/modules/balance"
	"github.com/Confialink/wallet-pkg-types"

	"github.com/shopspring/decimal"

	balanceModel "github.com/Confialink/wallet-accounts/internal/modules/balance/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	transactionModel "github.com/Confialink/wallet-accounts/internal/modules/transaction/model"
	userModel "github.com/Confialink/wallet-accounts/internal/modules/user/model"
)

const (
	// RateDesignationBaseReference indicates that request rate is specified for pair BaseCurrencyCode/ReferenceCurrencyCode
	RateDesignationBaseReference = "base/reference"
	// RateDesignationReferenceBase indicates that request rate is specified for pair ReferenceCurrencyCode/BaseCurrencyCode
	RateDesignationReferenceBase = "reference/base"
)

type Request struct {
	Id                    *uint64            `json:"id"`
	UserId                *string            `json:"userId"`
	User                  *RequestUser       `json:"user"`
	IsInitiatedByAdmin    *bool              `json:"isInitiatedByAdmin"`
	IsInitiatedBySystem   *bool              `json:"isInitiatedBySystem"`
	Status                *string            `json:"status"`
	Subject               *constants.Subject `json:"subject"`
	BaseCurrencyCode      *string            `json:"baseCurrencyCode"`
	ReferenceCurrencyCode *string            `json:"referenceCurrencyCode"`
	Rate                  *decimal.Decimal   `json:"rate"`
	// RateDesignation indicates how Rate is calculated
	// could be "base/reference" or "reference/base"
	RateDesignation string `json:"rateDesignation"`
	// Amount is always in currency that is specified by BaseCurrencyCode
	Amount *decimal.Decimal `json:"amount"`
	// InputAmount stores the value which was initially specified by a client
	// it is used in case when amount is specified in a reference currency (ReferenceCurrencyCode)
	// in this case Amount is calculated.
	// InputAmount is nil if the requested value is specified in a base currency (BaseCurrencyCode)
	InputAmount        *decimal.Decimal `json:"inputAmount"`
	Description        *string          `json:"description"`
	CancellationReason *string          `json:"cancellationReason"`
	CreatedAt          *time.Time       `json:"createdAt"`
	StatusChangedAt    *time.Time       `json:"statusChangedAt"`
	UpdatedAt          *time.Time
	Transactions       []*transactionModel.Transaction `gorm:"foreignkey:RequestId" json:"transactions"`
	IsVisible          *bool
	FullUser           *userModel.FullUser

	BalanceSnapshots []*balanceModel.Snapshot `json:"balanceSnapshots" gorm:"foreignkey:RequestId"`

	DataOwt *DataOwt `gorm:"foreignKey:RequestId"`

	Input types.DataJSON `json:"-"`

	BalanceDifference []*balance.Difference `json:"balanceDifference"`
}

type RequestUser struct {
	Id        *string `json:"id"`
	Username  *string `json:"username"`
	Email     *string `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

func (r *Request) GetInput() types.DataJSON {
	if r.Input == nil {
		r.Input = types.DataJSON{}
	}
	return r.Input
}

// SourceAccountId retrieves source account id from request input
// the second return value indicates whether it exists
func (r *Request) SourceAccountId() (int64, bool) {
	if id, ok := r.GetInput().Get("sourceAccountId"); ok {
		return conv.Int64FromInterface(id), true
	}
	return 0, false
}

// SourceAccountNumber retrieves source account number from request input
// the second return value indicates whether it exists
func (r *Request) SourceAccountNumber() (string, bool) {
	if number, ok := r.GetInput().Get("sourceAccountNumber"); ok {
		return number.(string), true
	}
	return "", false
}

// DestinationAccountId retrieves source account id from request input
// the second return value indicates whether it exists
func (r *Request) DestinationAccountId() (int64, bool) {
	if id, ok := r.GetInput().Get("destinationAccountId"); ok {
		return conv.Int64FromInterface(id), true
	}
	return 0, false
}

// DestinationAccountNumber retrieves destination account number from request input
// the second return value indicates whether it exists
func (r *Request) DestinationAccountNumber() (string, bool) {
	if number, ok := r.GetInput().Get("destinationAccountNumber"); ok {
		return number.(string), true
	}
	return "", false
}

// GetInputAmount returns requested amount based on rate designation
func (r *Request) GetInputAmount() decimal.Decimal {
	if r.RateDesignation == RateDesignationBaseReference {
		if r.Amount == nil {
			return decimal.NewFromInt(0)
		}
		return *r.Amount
	}
	if r.InputAmount == nil {
		return decimal.NewFromInt(0)
	}
	return *r.InputAmount
}

// RateBaseCurrencyCode returns rate base currency code (which depends on rate designation)
func (r *Request) RateBaseCurrencyCode() string {
	if r.RateDesignation == RateDesignationReferenceBase {
		return *r.ReferenceCurrencyCode
	}
	return *r.BaseCurrencyCode
}

// RateReferenceCurrencyCode returns rate reference currency code (which depends on rate designation)
func (r *Request) RateReferenceCurrencyCode() string {
	if r.RateDesignation == RateDesignationReferenceBase {
		return *r.BaseCurrencyCode
	}
	return *r.ReferenceCurrencyCode
}

// IsInitiatedByUser if request is initiated by user
func (r *Request) IsInitiatedByUser() bool {
	return !value.FromBool(r.IsInitiatedByAdmin) && !value.FromBool(r.IsInitiatedBySystem)
}

func (r *Request) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":                    r.Id,
		"userId":                r.UserId,
		"status":                r.Status,
		"subject":               r.Subject,
		"baseCurrencyCode":      r.BaseCurrencyCode,
		"referenceCurrencyCode": r.ReferenceCurrencyCode,
		"amount":                r.Amount,
		"description":           r.Description,
		"rate":                  r.Rate,
		"createdAt":             r.CreatedAt,
		"statusChangedAt":       r.StatusChangedAt,
		"updatedAt":             r.UpdatedAt,
		"cancellationReason":    r.CancellationReason,
		"isInitiatedBySystem":   r.IsInitiatedBySystem,
	})
}
