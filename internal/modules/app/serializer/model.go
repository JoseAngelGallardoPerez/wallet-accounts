package serializer

import (
	"reflect"
	"time"
)

type ModelSerializerInterface interface {
	Serialize(model interface{}, fields []interface{}) map[string]interface{}
	FilterFields(model interface{}, fields []string)
	FilterMapFields(mapData map[string]interface{}, fields []string)
	StrToTime(string) (time.Time, error)
	TimeToStr(time.Time) string
	GetNestedFieldsByName([]interface{}, string) []interface{}
}

type modelSerializer struct {
}

func NewModelSerializer() ModelSerializerInterface {
	return &modelSerializer{}
}

// Serialize serializes any model struct by passed fields map.
// Fields should equal names defined in struct
func (s *modelSerializer) Serialize(
	model interface{}, fields []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	modelValue := reflect.ValueOf(model).Elem()
	modelType := modelValue.Type()
	for _, name := range fields {
		if reflect.ValueOf(name).Type().Name() == "string" {
			fieldNameStr := name.(string)
			fieldValue := modelValue.FieldByName(fieldNameStr)
			fieldType, _ := modelType.FieldByName(fieldNameStr)
			serializedName := fieldType.Tag.Get("json")
			result[serializedName] = fieldValue.Interface()
		} else {
			for fieldNameStr, mapFields := range name.(map[string][]interface{}) {
				fieldPtr := modelValue.FieldByName(fieldNameStr)
				fieldType, _ := modelType.FieldByName(fieldNameStr)
				serializedName := fieldType.Tag.Get("json")
				result[serializedName] = s.Serialize(fieldPtr.Interface(), mapFields)
			}
		}
	}
	return result
}

func (s *modelSerializer) GetNestedFieldsByName(fields []interface{}, fieldName string) []interface{} {
	for _, name := range fields {
		if reflect.ValueOf(name).Type().Name() != "string" {
			nestedModel := name.(map[string][]interface{})

			for nestedName, nestedFields := range nestedModel {
				if nestedName == fieldName {
					return nestedFields
				}
			}
		}
	}
	return make([]interface{}, 0)
}

// FilterFields sets nil fot struct field if field is not in fields array
func (s *modelSerializer) FilterFields(
	model interface{}, fields []string) {

	modelValue := reflect.ValueOf(model).Elem()
	modelType := modelValue.Type()
	fieldsCount := modelValue.NumField()

	for i := 0; i < fieldsCount; i++ {
		field := modelValue.Field(i)
		if !field.IsNil() && !containsField(modelType.Field(i).Name, fields) {
			field.Set(reflect.Zero(field.Type()))
		}
	}
}

// FilterMapFields removes fields not in array and nils
func (s *modelSerializer) FilterMapFields(
	mapData map[string]interface{}, fields []string) {
	for k, v := range mapData {
		if !containsField(k, fields) || isNilInterface(v) {
			delete(mapData, k)
		}
	}
}

// StrToTime converts string time to time struct
func (s *modelSerializer) StrToTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// TimeToStr converts time to string in rfc3339 format in UTC
func (s *modelSerializer) TimeToStr(timeValue time.Time) string {
	utcLoc, _ := time.LoadLocation("UTC")
	return timeValue.In(utcLoc).Format(time.RFC3339)
}

func isNilInterface(v interface{}) bool {
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
		return true
	}
	return false
}

func containsField(field string, fields []string) bool {
	for _, v := range fields {
		if field == v {
			return true
		}
	}

	return false
}
