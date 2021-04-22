package import_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
)

type RequestDescriptionValidator struct {
	ImportCsvValidatorMain
}

func NewRequestDescriptionValidator() *RequestDescriptionValidator {
	return &RequestDescriptionValidator{}
}

func (v *RequestDescriptionValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.Description == "" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Description is required", row.Number)))
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestDescriptionValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
