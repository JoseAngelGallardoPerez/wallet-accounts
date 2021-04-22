package transfer

import "github.com/shopspring/decimal"

// CurrentBalance defines base operations
type Balance interface {
	// Amount shows current balance amount
	Amount() decimal.Decimal
	// Add adds value "v" to current balance
	Add(v decimal.Decimal) error
	// Sub subtracts value "v" from current balance
	Sub(v decimal.Decimal) error
}

// SimpleBalance is the most simple implementation of CurrentBalance which is a wrapper around decimal type
type SimpleBalance struct {
	v decimal.Decimal
}

// SimpleBalance constructor
func NewSimpleBalance(v decimal.Decimal) Balance {
	return &SimpleBalance{v: v}
}

// Amount shows current balance value
func (s *SimpleBalance) Amount() decimal.Decimal {
	return s.v
}

// Add adds value "v" to current balance
func (s *SimpleBalance) Add(v decimal.Decimal) error {
	s.v = s.v.Add(v)
	return nil
}

// Sub subtracts value "v" from current balance
func (s *SimpleBalance) Sub(v decimal.Decimal) error {
	s.v = s.v.Sub(v)
	return nil
}

// LinkedBalance is the most simple implementation of CurrentBalance which is a wrapper around decimal type
type LinkedBalance struct {
	v *decimal.Decimal
}

// LinkedBalance constructor
func NewLinkedBalance(v *decimal.Decimal) Balance {
	return &LinkedBalance{v: v}
}

// Amount shows current balance value
func (l *LinkedBalance) Amount() decimal.Decimal {
	return *l.v
}

// Add adds value "v" to current balance
func (l *LinkedBalance) Add(v decimal.Decimal) error {
	*l.v = l.v.Add(v)
	return nil
}

// Sub subtracts value "v" from current balance
func (l *LinkedBalance) Sub(v decimal.Decimal) error {
	*l.v = l.v.Sub(v)
	return nil
}
