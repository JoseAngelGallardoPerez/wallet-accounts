package serializer

import (
	"github.com/Confialink/wallet-accounts/internal/modules/currency/model"
	pb "github.com/Confialink/wallet-currencies/rpc/currencies"
)

type CurrencySerializerInterface interface {
	Deserialize(*pb.CurrencyResp) *model.Currency
	DeserializeList(*pb.CurrenciesResp) []*model.Currency
}

type currencySerializer struct {
}

func NewCurrencySerializer() CurrencySerializerInterface {
	return &currencySerializer{}
}

func (s *currencySerializer) Deserialize(response *pb.CurrencyResp) *model.Currency {
	return &model.Currency{
		Id:            response.Id,
		Code:          response.Code,
		Active:        response.Active,
		DecimalPlaces: int(response.DecimalPlaces),
	}
}

func (s *currencySerializer) DeserializeList(response *pb.CurrenciesResp) []*model.Currency {
	currencies := make([]*model.Currency, len(response.Currencies))
	for i, v := range response.Currencies {
		currencies[i] = s.Deserialize(v)
	}
	return currencies
}
