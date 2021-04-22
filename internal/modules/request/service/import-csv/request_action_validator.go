package import_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
)

const ActionCredit = "credit"
const ActionDebit = "debit"

type RequestActionValidator struct {
	ImportCsvValidatorMain
}

func NewRequestActionValidator() *RequestActionValidator {
	return &RequestActionValidator{}
}

func (v *RequestActionValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.Action == "" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Debit or CreditFromAlias field is required",
				row.Number)))
	} else if row.Action != ActionDebit && row.Action != ActionCredit {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Debit/CreditFromAlias is not valid",
				row.Number)))
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestActionValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
