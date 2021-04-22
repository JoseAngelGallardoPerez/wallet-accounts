package model

type Currency struct {
	Id            uint32 `json:"id"`
	Code          string `json:"code"`
	Active        bool
	DecimalPlaces int
}
