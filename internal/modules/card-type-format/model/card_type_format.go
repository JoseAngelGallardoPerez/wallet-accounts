package model

const CodeAlphanumeric = "alphanumeric"
const CodeSixteenNumeric = "sixteen_numeric"

var Rules = map[string]string{
	CodeSixteenNumeric: "^[0-9]{16}$",
	CodeAlphanumeric:   "^[a-zA-Z0-9-*]{1,20}$",
}

type CardTypeFormat struct {
	Id   *uint32 `json:"id"`
	Name *string `json:"name"`
	Code *string `json:"code"`
}
