package serializer

import (
	"github.com/Confialink/wallet-accounts/internal/modules/account/model"
	"github.com/Confialink/wallet-pkg-model_serializer"
)

type AccountSerializerInterface interface {
	Serialize(account *model.Account, fields []interface{}) map[string]interface{}
	SerializeList(accounts []*model.Account, fields []interface{}) []map[string]interface{}
}

type accountSerializer struct{}

func NewAccountSerializer() AccountSerializerInterface {
	return &accountSerializer{}
}

// Serialize transforms Account model into map with fields
func (self *accountSerializer) Serialize(account *model.Account, fields []interface{}) map[string]interface{} {
	return model_serializer.Serialize(account, fields)
}

func (self *accountSerializer) SerializeList(accounts []*model.Account, fields []interface{}) []map[string]interface{} {
	serialized := make([]map[string]interface{}, len(accounts))

	for i, v := range accounts {
		serialized[i] = self.Serialize(v, fields)
	}
	return serialized
}
