package serializer

import (
	"encoding/json"

	modelSerializer "github.com/Confialink/wallet-accounts/internal/modules/app/serializer"
	cardTypeCategory "github.com/Confialink/wallet-accounts/internal/modules/card-type-category/model"
	cardTypeFormat "github.com/Confialink/wallet-accounts/internal/modules/card-type-format/model"
	"github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	"github.com/fatih/structs"
)

type CardTypeSerializerInterface interface {
	Serialize(cardType *model.CardType, fields []interface{}) map[string]interface{}
	SerializeToModel(cardType *model.CardType) *model.SerializedCardType
	SerializeList(cardTypes []*model.CardType, fields []interface{}) []map[string]interface{}
	Deserialize(data *[]byte, fields []string) (*model.SerializedCardType, error)
	DeserializeFields(data *[]byte, fields []string) (map[string]interface{}, error)
}

type cardTypeSerializer struct {
	modelSerializer.ModelSerializerInterface
}

func NewCardTypeSerializer(
	modelSerializer modelSerializer.ModelSerializerInterface,
) CardTypeSerializerInterface {
	return &cardTypeSerializer{modelSerializer}
}

func (s *cardTypeSerializer) Serialize(
	cardType *model.CardType, fields []interface{}) map[string]interface{} {
	serializedModel := s.SerializeToModel(cardType)
	return s.ModelSerializerInterface.Serialize(serializedModel, fields)
}

func (s *cardTypeSerializer) SerializeToModel(
	cardType *model.CardType) *model.SerializedCardType {
	var serializedCardCategory cardTypeCategory.CardTypeCategory
	if cardType.Category != nil {
		serializedCardCategory = cardTypeCategory.CardTypeCategory{Id: cardType.Category.Id, Name: cardType.Category.Name}
	}

	var serializedCardFormat cardTypeFormat.CardTypeFormat
	if cardType.Format != nil {
		serializedCardFormat = cardTypeFormat.CardTypeFormat{Id: cardType.Format.Id, Name: cardType.Format.Name, Code: cardType.Format.Code}
	}

	return &model.SerializedCardType{
		Id:                 cardType.Id,
		Name:               cardType.Name,
		CurrencyCode:       cardType.CurrencyCode,
		IconId:             cardType.IconId,
		CardTypeCategoryId: cardType.CardTypeCategoryId,
		CardTypeFormatId:   cardType.CardTypeFormatId,
		Category:           &serializedCardCategory,
		Format:             &serializedCardFormat,
	}
}

// Serializes list with Serialize method
func (s *cardTypeSerializer) SerializeList(
	cardTypes []*model.CardType, fields []interface{}) []map[string]interface{} {
	result := make([]map[string]interface{}, len(cardTypes))
	for i := range result {
		result[i] = s.Serialize(cardTypes[i], fields)
	}
	return result
}

func (s *cardTypeSerializer) Deserialize(
	data *[]byte, fields []string) (*model.SerializedCardType, error) {
	serializedData := model.SerializedCardType{}
	if err := json.Unmarshal(*data, &serializedData); err != nil {
		return nil, err
	}

	s.FilterFields(&serializedData, fields)
	return &serializedData, nil
}

// Deserialize data from request into map with only passed fields
func (s *cardTypeSerializer) DeserializeFields(
	data *[]byte, fields []string) (map[string]interface{}, error) {
	serializedCardType := model.SerializedCardType{}
	if err := json.Unmarshal(*data, &serializedCardType); err != nil {
		return nil, err
	}
	serializedMap := structs.Map(&serializedCardType)
	s.FilterMapFields(serializedMap, fields)
	return serializedMap, nil
}
