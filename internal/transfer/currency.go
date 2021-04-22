package transfer

import "github.com/pkg/errors"

// Currency represents money currency information required for transfers
type Currency struct {
	code     string
	fraction uint
}

// NewCurrency is Currency constructor
func NewCurrency(code string, fraction uint) *Currency {
	return &Currency{code: code, fraction: fraction}
}

// Fraction represents a fraction of a currency unit (number of decimal places) e.g. 2 for USD - 10^2 (100 cents)
func (c *Currency) Fraction() uint {
	return c.fraction
}

// Code returns currency code i.e. EUR, USD etc.
func (c *Currency) Code() string {
	return c.code
}

// String returns currency string representation
func (c *Currency) String() string {
	return c.code
}

// CurrencyProvider defines currency source
type CurrencyProvider interface {
	Get(code string) (Currency, error)
}

// CurrencyConsumer defines instances that can accept currencies
type CurrencyConsumer interface {
	Add(currency Currency) error
}

// CurrencyBox defines instances that can both consume and provide currencies
type CurrencyBox interface {
	CurrencyConsumer
	CurrencyProvider
}

// DirectCurrencySource is used as a simple currencies container
type DirectCurrencySource map[string]Currency

// NewDirectCurrencySource initializes DirectCurrencySource
func NewDirectCurrencySource() DirectCurrencySource {
	return make(DirectCurrencySource)
}

// Add adds new currency
func (d DirectCurrencySource) Add(currency Currency) error {
	d[currency.code] = currency
	return nil
}

// Get retrieves added currency
func (d DirectCurrencySource) Get(code string) (Currency, error) {
	if curr, ok := d[code]; ok {
		return curr, nil
	}
	return Currency{}, errors.Wrapf(ErrCurrencyNotFound, "currency %s is not found", code)
}
