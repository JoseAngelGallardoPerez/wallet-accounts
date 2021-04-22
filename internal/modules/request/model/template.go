package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-pkg-utils/pointer"
)

type Template struct {
	Id             *uint64            `json:"id"`
	Name           *string            `json:"name"`
	RequestSubject *constants.Subject `json:"requestSubject"`
	UserId         *string            `json:"userId"`
	CreatedAt      *time.Time         `json:"createdAt"`
	Data           *string            `json:"_"`
}

func (t *Template) SetData(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.Data = pointer.ToString(string(bytes))
	return nil
}

func (t *Template) Unmarshal(v interface{}) error {
	if t.Data == nil {
		return errors.New("data is nil")
	}

	return json.Unmarshal([]byte(*t.Data), v)
}

type Templates []*Template

func (t Templates) Unmarshal(v interface{}) error {
	strJson := "["
	for _, template := range t {
		if template.Data == nil {
			return errcodes.CreatePublicError(errcodes.CodeInvalidTemplate, "some of templates has data field set to nil")
		}
		strJson += *template.Data + ","
	}
	strJson = strings.TrimRight(strJson, ",") + "]"
	return json.Unmarshal([]byte(strJson), v)
}

func (t *Template) TableName() string {
	return "request_templates"
}

func (t *Template) MarshalJSON() ([]byte, error) {
	var data map[string]interface{}
	err := t.Unmarshal(&data)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":             t.Id,
		"name":           t.Name,
		"requestSubject": t.RequestSubject,
		"data":           data,
		"createdAt":      t.CreatedAt,
	}
	return json.Marshal(result)
}
