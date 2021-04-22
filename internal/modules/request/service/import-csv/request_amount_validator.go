package import_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
)

type RequestAmountValidator struct {
	ImportCsvValidatorMain
}

func NewRequestAmountValidator() *RequestAmountValidator {
	return &RequestAmountValidator{}
}

func (v *RequestAmountValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.Amount == "" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Amount field is required",
				row.Number)))
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestAmountValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
