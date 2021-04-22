package serializer

import (
	requestModel "github.com/Confialink/wallet-accounts/internal/modules/request/model"
	"github.com/Confialink/wallet-pkg-model_serializer"
)

const balanceDifferenceFieldName = "balanceDifference"

func ProvideBalanceDifferenceSerializer() model_serializer.FieldSerializer {
	return func(model interface{}) (fieldName string, value interface{}) {
		request := model.(*requestModel.Request)

		return balanceDifferenceFieldName, request.BalanceDifference
	}
}
