package import_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
)

type RequestApplyIwtFeeValidator struct {
	ImportCsvValidatorMain
}

func NewRequestApplyIwtFeeValidator() *RequestApplyIwtFeeValidator {
	return &RequestApplyIwtFeeValidator{}
}

func (v *RequestApplyIwtFeeValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.ApplyIwtFee != "yes" && row.ApplyIwtFee != "no" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Apply IWT fee field is required",
				row.Number)))
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestApplyIwtFeeValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
