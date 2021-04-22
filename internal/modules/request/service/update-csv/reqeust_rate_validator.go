package update_csv

import (
	"fmt"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/shopspring/decimal"
)

type RequestRateValidator struct {
	UpdateCsvValidatorMain
}

func NewRequestRateValidator() *RequestRateValidator {
	return &RequestRateValidator{}
}

func (v *RequestRateValidator) Validate(row Row, c *Context) []errors.TypedError {
	var rate decimal.Decimal
	var err error
	request := c.GetRequest()

	if row.Len == 3 && row.Rate != "" {
		rate, err = decimal.NewFromString(row.Rate)
		if nil != err {
			c.AddError(errcodes.CreatePublicError(errcodes.CodeCsvFileInvalidRow,
				fmt.Sprintf("Row number %d of CSV file failed. Reason - Rate is invalid",
					row.Number)))
		}
	} else if request != nil {
		rate = *request.Rate
	}

	c.SetRate(rate)

	if v.GetNext() != nil {
		v.GetNext().Validate(row, c)
	}

	return c.GetErrors()
}

func (v *RequestRateValidator) SetNext(next UpdateCsvValidator) UpdateCsvValidator {
	v.next = next
	return next
}
