package import_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	"github.com/Confialink/wallet-pkg-errors"
)

type RequestAccountNumberValidator struct {
	ImportCsvValidatorMain
	accRepo *repository.AccountRepository
}

func NewRequestAccountNumberValidator(accRepo *repository.AccountRepository) *RequestAccountNumberValidator {
	return &RequestAccountNumberValidator{
		accRepo: accRepo,
	}
}

func (v *RequestAccountNumberValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.AccountNumber == "" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Account Number is required",
				row.Number)))
	} else {
		fmt.Println(row.AccountNumber)
		_, err := v.accRepo.FindByNumber(row.AccountNumber)
		if err != nil {
			c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
				fmt.Sprintf("Row number %d of CSV file failed. Reason - Account Number is not valid",
					row.Number)))
		}
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestAccountNumberValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
