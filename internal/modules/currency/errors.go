package currency

import "github.com/Confialink/wallet-accounts/internal/errcodes"

type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

const (
	ErrExchangeRateNotFound = Error(errcodes.CodeExchangeRateNotFound)
)
