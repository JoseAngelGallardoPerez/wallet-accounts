package update_csv

import (
	"fmt"
	"strconv"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
	"github.com/Confialink/wallet-pkg-errors"
)

type RequestExistValidator struct {
	UpdateCsvValidatorMain
	repo repository.RequestRepositoryInterface
}

func NewRequestExistValidator(repo repository.RequestRepositoryInterface) *RequestExistValidator {
	return &RequestExistValidator{
		repo: repo,
	}
}

func (v *RequestExistValidator) Validate(row Row, c *Context) []errors.TypedError {
	if row.RequestId == "" {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Request ID is required",
				row.Number)))
	} else {
		requestId, _ := strconv.ParseUint(row.RequestId, 10, 64)
		result, _ := v.repo.FindById(requestId)

		if nil == result {
			c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
				fmt.Sprintf("Row number %d of CSV file failed. Reason - Request ID is not valid",
					row.Number)))
		}

		c.SetRequest(result)
	}

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestExistValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
