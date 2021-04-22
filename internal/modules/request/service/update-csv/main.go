package update_csv

import (
	"github.com/Confialink/wallet-pkg-errors"
)

type UpdateCsvValidator interface {
	SetNext(next UpdateCsvValidator) UpdateCsvValidator
	GetNext() UpdateCsvValidator
	Validate(row Row, c *Context) []errors.TypedError
}

type UpdateCsvValidatorMain struct {
	prev      UpdateCsvValidator
	next      UpdateCsvValidator
	context   *Context
	rowNumber uint64
}

func (v *UpdateCsvValidatorMain) GetNext() UpdateCsvValidator {
	return v.next
}

func (v *UpdateCsvValidatorMain) SetPrev(prev UpdateCsvValidator) {
	v.prev = prev
}
