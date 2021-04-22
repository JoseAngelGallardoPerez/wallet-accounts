package import_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
)

type RequestRevenueValidator struct {
	ImportCsvValidatorMain
}

func NewRequestRevenueValidator() *RequestRevenueValidator {
	return &RequestRevenueValidator{}
}

func (v *RequestRevenueValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.Revenue != "yes" && row.Revenue != "no" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Revenue (yes/no) field is required",
				row.Number)))
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestRevenueValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
