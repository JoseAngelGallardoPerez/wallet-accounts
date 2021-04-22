package calculation

import (
	"fmt"
	"time"

	"github.com/jinzhu/now"
	"github.com/shopspring/decimal"
)

// annualInterestForNDays calculates interest value by number of days
func annualInterestForNDays(amount decimal.Decimal, annualFeePercent decimal.Decimal, numberOfDays uint, year ...int) decimal.Decimal {
	forYear := time.Now().Year()
	if len(year) > 0 {
		forYear = year[0]
	}
	layout := "2006-01-02"
	format := "%d-01-01"
	withYear, _ := time.Parse(layout, fmt.Sprintf(format, forYear))

	daysInYear := decimal.New(int64(now.New(withYear).EndOfYear().Sub(now.New(withYear).BeginningOfYear()).Hours()/24), 0)

	oneHundred := decimal.New(100, 0)
	nDays := decimal.New(int64(numberOfDays), 0)
	// amount * (percent / 100) / daysInYear * numberOfDays
	return amount.Mul(annualFeePercent).Div(oneHundred).Div(daysInYear).Mul(nDays)
}
