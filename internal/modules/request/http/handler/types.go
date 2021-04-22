package handler

import (
	"encoding/json"

	"github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/types"
)

type recipient struct {
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}

type preview struct {
	IncomingAmount       string        `json:"incomingAmount,omitempty"`
	IncomingAmountInFiat string        `json:"incomingAmountInFiat,omitempty"`
	IncomingCurrencyCode string        `json:"incomingCurrencyCode,omitempty"`
	TotalOutgoingAmount  string        `json:"totalOutgoingAmount,omitempty"`
	Details              types.Details `json:"details"`
	Recipient            *recipient
}

func (p *preview) MarshalJSON() ([]byte, error) {
	detailsToShow := make([]interface{}, 0, 1)
	showableDetails := p.Details.ByPurposes(
		constants.PurposeFeeTransfer,
		constants.PurposeFeeExchangeMargin,
	)
	for _, detail := range showableDetails {
		fields := map[string]interface{}{
			"purpose":      detail.Purpose.String(),
			"amount":       detail.Amount,
			"currencyCode": detail.CurrencyCode,
		}
		detailsToShow = append(detailsToShow, fields)
	}

	obj := map[string]interface{}{
		"details": detailsToShow,
	}
	if p.TotalOutgoingAmount != "" {
		obj["totalOutgoingAmount"] = p.TotalOutgoingAmount
	}
	if p.IncomingAmount != "" {
		obj["incomingAmount"] = p.IncomingAmount
	}
	if p.IncomingAmountInFiat != "" {
		obj["incomingAmountInFiat"] = p.IncomingAmountInFiat
	}
	if p.IncomingCurrencyCode != "" {
		obj["incomingCurrencyCode"] = p.IncomingCurrencyCode
	}
	return json.Marshal(obj)
}
