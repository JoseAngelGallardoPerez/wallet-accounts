package import_csv

import (
	"github.com/Confialink/wallet-pkg-errors"
)

type UpdateCsvValidator interface {
	SetNext(next UpdateCsvValidator) UpdateCsvValidator
	GetNext() UpdateCsvValidator
	Validate(row Row, c *Context) []errors.TypedError
}

type ImportCsvValidatorMain struct {
	prev      UpdateCsvValidator
	next      UpdateCsvValidator
	context   *Context
	rowNumber uint64
}

func (v *ImportCsvValidatorMain) GetNext() UpdateCsvValidator {
	return v.next
}

func (v *ImportCsvValidatorMain) SetPrev(prev UpdateCsvValidator) {
	v.prev = prev
}
