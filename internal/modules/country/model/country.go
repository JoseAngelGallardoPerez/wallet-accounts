package model

type Country struct {
	Id          *uint   `gorm:"primary_key" json:"id"`
	Name        *string `json:"name"`
	Code        *string `json:"code"`
	Code3       *string `json:"code3"`
	CodeNumeric *string `json:"codeNumeric"`
}
