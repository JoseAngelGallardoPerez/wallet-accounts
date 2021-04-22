package update_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
)

type LineLengthValidator struct {
	UpdateCsvValidatorMain
}

func NewLineNengthValidator() *LineLengthValidator {
	return &LineLengthValidator{}
}

func (v *LineLengthValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.Len != 3 && row.Len != 2 {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Invalid number of columns. Expected 2 or 3, got %d",
				row.Number,
				row.Len)))

		// we must return errors if line length is not valid
		return c.GetErrors()
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *LineLengthValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
