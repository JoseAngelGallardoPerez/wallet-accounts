package serializer

import (
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	txConstants "github.com/Confialink/wallet-accounts/internal/modules/transaction/constants"
	"github.com/Confialink/wallet-pkg-model_serializer"
)

const amountFieldName = "amount"

func ProvideAmountSerializer() model_serializer.FieldSerializer {
	return func(model interface{}) (fieldName string, value interface{}) {
		request := model.(*requestModel.Request)
		if *request.Subject != constants.SubjectTransferOutgoingWireTransfer {
			return amountFieldName, request.Amount
		}

		amount := *request.Amount
		for _, tx := range request.Transactions {
			if *tx.Purpose == txConstants.PurposeFeeExchangeMargin.String() {
				amount = amount.Add(tx.Amount.Abs())
			}
		}

		return amountFieldName, amount
	}
}
