package calculation

import (
	"github.com/shopspring/decimal"
	"github.com/Confialink/wallet-pkg-errors"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/currency/service"
	"github.com/Confialink/wallet-accounts/internal/modules/money"
)

const DBDecimalPrecision = 18

var DefaultCrossRateRounder CrossRateRounder = DBDecimalPrecisionRounding

type CrossRateRounder func(value decimal.Decimal) decimal.Decimal

func DBDecimalPrecisionRounding(value decimal.Decimal) decimal.Decimal {
	return value.Round(DBDecimalPrecision)
}

type AmountRounder func(amount money.Amount) (money.Amount, errors.TypedError)

type Rounding struct {
	currenciesService service.CurrenciesServiceInterface
}

func NewRounding(currenciesService service.CurrenciesServiceInterface) *Rounding {
	return &Rounding{currenciesService: currenciesService}
}

//TruncateAmount truncates the amount by passed currency precision
func (r *Rounding) TruncateAmount(input money.Amount) (money.Amount, errors.TypedError) {
	currency, err := r.currenciesService.GetByCode(input.CurrencyCode)
	if err != nil {
		return input, &errors.PrivateError{OriginalError: err}
	}

	return money.Amount{
		Value:        input.Value.Truncate(int32(currency.DecimalPlaces)),
		CurrencyCode: input.CurrencyCode,
	}, nil
}

//RoundAmount rounds the amount by passed currency precision with bank method
func (r *Rounding) RoundAmount(input money.Amount) (money.Amount, errors.TypedError) {
	currency, err := r.currenciesService.GetByCode(input.CurrencyCode)
	if err != nil {
		return input, &errors.PrivateError{OriginalError: err}
	}

	return money.Amount{
		Value:        input.Value.RoundBank(int32(currency.DecimalPlaces)),
		CurrencyCode: input.CurrencyCode,
	}, nil
}

func (r *Rounding) ValidatePrecision(input money.Amount) errors.TypedError {
	truncated, err := r.TruncateAmount(input)
	if err != nil {
		return err
	}

	if !truncated.Value.Equal(input.Value) {
		return errcodes.CreatePublicError(errcodes.CodeInvalidCurrencyPrecision)
	}

	return nil
}
