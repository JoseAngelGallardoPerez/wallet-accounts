package update_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/request/constants"
	"github.com/Confialink/wallet-pkg-errors"
)

const StatusExecuted = "executed"
const StatusCancelled = "cancelled"
const StatusCanceled = "canceled"

type RequestStatusValidator struct {
	UpdateCsvValidatorMain
}

func NewRequestStatusValidator() *RequestStatusValidator {
	return &RequestStatusValidator{}
}

func (v *RequestStatusValidator) Validate(row Row, c *Context) []errors.TypedError {
	request := c.GetRequest()
	if request != nil && request.Status != nil && *request.Status != constants.StatusPending {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Request %d is already executed or cancelled",
				row.Number,
				*request.Id)))
	}

	if row.Status != StatusExecuted && row.Status != StatusCancelled {
		c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
			fmt.Sprintf("Row number %d of CSV file failed. Reason - Status is invalid",
				row.Number)))
	}

	c.SetStatus(row.Status)

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestStatusValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
