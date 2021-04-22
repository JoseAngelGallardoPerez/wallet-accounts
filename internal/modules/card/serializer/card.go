package serializer

import (
	"encoding/json"

	"github.com/fatih/structs"

	modelSerializer "github.com/Confialink/wallet-accounts/internal/modules/app/serializer"
	cardTypeModel "github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	cardTypeSerializer "github.com/Confialink/wallet-accounts/internal/modules/card-type/serializer"
	"github.com/Confialink/wallet-accounts/internal/modules/card/model"
)

type CardSerializerInterface interface {
	Serialize(card *model.Card, fields []interface{}) map[string]interface{}
	Deserialize(data *[]byte, fields []string) (*model.SerializedCard, error)
	DeserializeFields(data *[]byte, fields []string) (map[string]interface{}, error)
}

type cardSerializer struct {
	modelSerializer    modelSerializer.ModelSerializerInterface
	cardTypeSerializer cardTypeSerializer.CardTypeSerializerInterface
}

func NewCardSerializer(
	modelSerializer modelSerializer.ModelSerializerInterface,
	cardTypeSerializer cardTypeSerializer.CardTypeSerializerInterface,
) CardSerializerInterface {
	return &cardSerializer{modelSerializer, cardTypeSerializer}
}

// Serialize transforms Card model into map with fields
func (s *cardSerializer) Serialize(card *model.Card, fields []interface{}) map[string]interface{} {
	createdAt := s.modelSerializer.TimeToStr(*(card.CreatedAt))
	var serializedCardType *cardTypeModel.SerializedCardType
	if card.CardType != nil {
		serializedCardType = s.cardTypeSerializer.SerializeToModel(card.CardType)
	}
	serializedModel := &model.SerializedCard{
		Id:              card.Id,
		Number:          card.Number,
		Status:          card.Status,
		Balance:         card.Balance,
		CardTypeId:      card.CardTypeId,
		UserId:          card.UserId,
		ExpirationYear:  card.ExpirationYear,
		ExpirationMonth: card.ExpirationMonth,
		CreatedAt:       &createdAt,
		CardType:        serializedCardType,
		User:            card.User,
	}
	return s.modelSerializer.Serialize(serializedModel, fields)
}

// Deserialize transforms json data into serialized card
func (s *cardSerializer) Deserialize(data *[]byte, fields []string) (*model.SerializedCard, error) {
	serializedData := model.SerializedCard{}
	if err := json.Unmarshal(*data, &serializedData); err != nil {
		return nil, err
	}

	s.modelSerializer.FilterFields(&serializedData, fields)
	return &serializedData, nil
}

// DeserializeFields converts json data into map fields with transformed values
func (s *cardSerializer) DeserializeFields(data *[]byte, fields []string) (map[string]interface{}, error) {
	serializedCard := model.SerializedCard{}
	if err := json.Unmarshal(*data, &serializedCard); err != nil {
		return nil, err
	}
	serializedMap := structs.Map(&serializedCard)
	s.modelSerializer.FilterMapFields(serializedMap, fields)

	return serializedMap, nil
}
