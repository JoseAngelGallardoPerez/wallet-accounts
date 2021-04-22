package model

import (
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"regexp"
	"time"

	"encoding/json"

	accountModel "github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/shopspring/decimal"
)

const (
	StatusPending   = "pending"
	StatusExecuted  = "executed"
	StatusCancelled = "cancelled"
)
const (
	TypeAccount = "account"
	TypeCard    = "card"
	TypeFee     = "fee"
	TypeRevenue = "revenue"
)

var (
	incomingExpr = regexp.MustCompile(`^\w{2,3}_incoming$`)
	outgoingExpr = regexp.MustCompile(`^\w{2,3}_outgoing$`)
)

type Transaction struct {
	Id                           *uint64               `json:"id"`
	RequestId                    *uint64               `json:"requestId"`
	AccountId                    *uint64               `json:"accountId"`
	Account                      *accountModel.Account `gorm:"foreignkey:AccountId;association_foreignkey:ID" json:"account"`
	CardId                       *uint32               `json:"cardId"`
	RevenueAccountId             *uint64
	RevenueAccount               *accountModel.RevenueAccountModel `gorm:"foreignkey:RevenueAccountId;association_foreignkey:ID" json:"revenueAccount"`
	Status                       *string                           `json:"status"`
	Description                  *string                           `json:"description"`
	Amount                       *decimal.Decimal                  `json:"amount"`
	ShowAmount                   *decimal.Decimal                  `json:"showAmount"`
	AvailableBalanceSnapshot     *decimal.Decimal                  `json:"availableBalanceSnapshot"`
	ShowAvailableBalanceSnapshot *decimal.Decimal                  `json:"showAvailableBalanceSnapshot"`
	IsVisible                    *bool                             `json:"isVisible" gorm:"default:'1'"`
	CurrentBalanceSnapshot       *decimal.Decimal                  `json:"currentBalanceSnapshot"`
	// there might be a case when outgoing transaction is executed before margin fee (OWT)
	// in this case since margin fee is hidden then outgoing transaction should also reflect margin amount
	ShowCurrentBalanceSnapshot *decimal.Decimal `json:"showCurrentBalanceSnapshot"`
	Type                       *string          `json:"type"`
	Purpose                    *string          `json:"purpose"`
	CreatedAt                  *time.Time       `json:"createdAt"`
	UpdatedAt                  *time.Time
}

func (t *Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) IsCredit() bool {
	return t.Amount != nil && t.Amount.GreaterThan(decimal.Zero)
}

func (t *Transaction) IsDebit() bool {
	return t.Amount != nil && t.Amount.LessThan(decimal.Zero)
}

func (t *Transaction) IsTargetOutgoing() bool {
	return t.Purpose != nil && outgoingExpr.MatchString(*t.Purpose)
}

func (t *Transaction) IsIncoming() bool {
	return t.Purpose != nil && incomingExpr.MatchString(*t.Purpose)
}

func (t *Transaction) IsExchangeMarginFee() bool {
	return t.Purpose != nil && *t.Purpose == constants.PurposeFeeExchangeMargin.String()
}

func (t *Transaction) IsDefaultTransferFee() bool {
	return t.Purpose != nil && *t.Purpose == constants.PurposeFeeTransfer.String()
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":               t.Id,
		"requestId":        t.RequestId,
		"status":           t.Status,
		"accountId":        t.AccountId,
		"account":          t.Account,
		"cardId":           t.CardId,
		"revenueAccountId": t.RevenueAccountId,
		"revenueAccount":   t.RevenueAccount,
		"description":      t.Description,
		"amount":           t.Amount,
		"balanceSnapshot":  t.AvailableBalanceSnapshot,
		"type":             t.Type,
		"purpose":          t.Purpose,
		"createdAt":        t.CreatedAt,
		"updatedAt":        t.UpdatedAt,
	})
}
