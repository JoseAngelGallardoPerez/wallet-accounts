package model

import (
	"encoding/json"
	"time"
)

type Settings struct {
	Id          uint32
	Name        string
	Value       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (s *Settings) IsExist() bool {
	return s.Id > 0
}

func (s *Settings) MarshalJSON() ([]byte, error) {
	obj := map[string]interface{}{
		"type":        "setting",
		"name":        s.Name,
		"value":       s.Value,
		"description": s.Description,
	}

	return json.Marshal(obj)
}
