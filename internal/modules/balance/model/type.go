package model

type Type struct {
	Id   uint64 `json:"-"`
	Name string `json:"name"`
}

func (t *Type) TableName() string {
	return "balance_types"
}
