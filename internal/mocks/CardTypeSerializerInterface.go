// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import (
	model "github.com/Confialink/wallet-accounts/internal/modules/card-type/model"
	mock "github.com/stretchr/testify/mock"
)

// CardTypeSerializerInterface is an autogenerated mock type for the CardTypeSerializerInterface type
type CardTypeSerializerInterface struct {
	mock.Mock
}

// Deserialize provides a mock function with given fields: data, fields
func (_m *CardTypeSerializerInterface) Deserialize(data *[]byte, fields []string) (*model.SerializedCardType, error) {
	ret := _m.Called(data, fields)

	var r0 *model.SerializedCardType
	if rf, ok := ret.Get(0).(func(*[]byte, []string) *model.SerializedCardType); ok {
		r0 = rf(data, fields)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SerializedCardType)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*[]byte, []string) error); ok {
		r1 = rf(data, fields)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeserializeFields provides a mock function with given fields: data, fields
func (_m *CardTypeSerializerInterface) DeserializeFields(data *[]byte, fields []string) (map[string]interface{}, error) {
	ret := _m.Called(data, fields)

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func(*[]byte, []string) map[string]interface{}); ok {
		r0 = rf(data, fields)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*[]byte, []string) error); ok {
		r1 = rf(data, fields)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Serialize provides a mock function with given fields: cardType, fields
func (_m *CardTypeSerializerInterface) Serialize(cardType *model.CardType, fields []interface{}) map[string]interface{} {
	ret := _m.Called(cardType, fields)

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func(*model.CardType, []interface{}) map[string]interface{}); ok {
		r0 = rf(cardType, fields)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	return r0
}

// SerializeList provides a mock function with given fields: cardTypes, fields
func (_m *CardTypeSerializerInterface) SerializeList(cardTypes []*model.CardType, fields []interface{}) []map[string]interface{} {
	ret := _m.Called(cardTypes, fields)

	var r0 []map[string]interface{}
	if rf, ok := ret.Get(0).(func([]*model.CardType, []interface{}) []map[string]interface{}); ok {
		r0 = rf(cardTypes, fields)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]map[string]interface{})
		}
	}

	return r0
}
