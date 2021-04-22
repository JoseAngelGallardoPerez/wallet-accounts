package calculation

import (
	"errors"

	"github.com/shopspring/decimal"
)

type InterestCalculationMethod string

const InterestCalculationMethodDaily = InterestCalculationMethod("daily")

var knownMethods = map[string]InterestCalculationMethod{
	string(InterestCalculationMethodDaily): InterestCalculationMethodDaily,
}

func (i InterestCalculationMethod) Is(method string) bool {
	return method == string(i)
}

func MethodFromString(method string) (InterestCalculationMethod, error) {
	if result, ok := knownMethods[method]; ok {
		return result, nil
	}
	return InterestCalculationMethod(""), errors.New("unknown method " + method)
}

func (i InterestCalculationMethod) AnnualInterest(amount decimal.Decimal, annualFeePercent decimal.Decimal, year ...int) decimal.Decimal {
	switch i {
	case InterestCalculationMethodDaily:
		return annualInterestForNDays(amount, annualFeePercent, 1, year...)
	}
	panic("unresolved method")
}
