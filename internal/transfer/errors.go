package transfer

// Error defines rate error
type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

const (
	ErrNotEnoughFunds     = Error("not enough funds")
	ErrCurrenciesMismatch = Error("mismatch of currencies")
	ErrCurrencyNotFound   = Error("currency not found")
	ErrInvalidAmount      = Error("invalid amount")
	ErrAlreadyPerformed   = Error("action is already performed")
)
