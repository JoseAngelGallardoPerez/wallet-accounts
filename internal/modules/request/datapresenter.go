package request

import (
	"github.com/Confialink/wallet-accounts/internal/conv"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-pkg-utils/pointer"
)

// DataPresenter is aimed to provide request metadata such as source account, fee options etc.
type DataPresenter interface {
	// Present returns metadata
	Present(request *model.Request) (interface{}, error)
}

func NewDataPresenter(owtRepository *repository.DataOwt) DataPresenter {
	return &defaultDataPresenter{
		owtRepository: owtRepository,
	}
}

type defaultDataPresenter struct {
	owtRepository *repository.DataOwt
}

func (d *defaultDataPresenter) Present(request *model.Request) (interface{}, error) {
	if request.Subject.EqualsTo(constants.SubjectTransferOutgoingWireTransfer) {
		return d.owtRepository.FindByRequestId(*request.Id)
	}
	data := &requestData{}
	data.setFieldsFromMap(request.GetInput())
	return data, nil
}

// requestData represents possible options related to request
// it is used in order to "omitempty" is a simple way to get rid of unused option
type requestData struct {
	RevenueAccountId        *int64 `json:"revenueAccountId,omitempty"`
	SourceAccountId         *int64 `json:"sourceAccountId,omitempty"`
	DestinationAccountId    *int64 `json:"destinationAccountId,omitempty"`
	DestinationCardId       *int64 `json:"destinationCardId,omitempty"`
	DebitFromRevenueAccount *bool  `json:"debitFromRevenueAccount,omitempty"`
	CreditToRevenueAccount  *bool  `json:"creditToRevenueAccount,omitempty"`
	ApplyIwtFee             *bool  `json:"applyIwtFee,omitempty"`
}

func (r *requestData) setFieldsFromMap(m map[string]interface{}) {
	if v, ok := m["revenueAccountId"]; ok {
		r.RevenueAccountId = pointer.ToInt64(conv.Int64FromInterface(v))
	}
	if v, ok := m["sourceAccountId"]; ok {
		r.SourceAccountId = pointer.ToInt64(conv.Int64FromInterface(v))
	}
	if v, ok := m["destinationAccountId"]; ok {
		r.DestinationAccountId = pointer.ToInt64(conv.Int64FromInterface(v))
	}
	if v, ok := m["destinationCardId"]; ok {
		r.DestinationCardId = pointer.ToInt64(conv.Int64FromInterface(v))
	}
	if v, ok := m["debitFromRevenueAccount"]; ok {
		r.DebitFromRevenueAccount = pointer.ToBool(v.(bool))
	}
	if v, ok := m["creditToRevenueAccount"]; ok {
		r.CreditToRevenueAccount = pointer.ToBool(v.(bool))
	}
	if v, ok := m["applyIwtFee"]; ok {
		r.ApplyIwtFee = pointer.ToBool(v.(bool))
	}
}
