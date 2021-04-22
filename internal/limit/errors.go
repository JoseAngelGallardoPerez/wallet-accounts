package limit

import "github.com/Confialink/wallet-accounts/internal/errcodes"

// Error defines string error
type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

const (
	ErrLimitExceeded      = Error(errcodes.CodeLimitExceeded)
	ErrInvalidAmount      = Error("invalid amount given")
	ErrAlreadyExist       = Error("limit is already exist")
	ErrIdIncomplete       = Error("one or more identifier properties are missed")
	ErrNotFound           = Error("limit is not found")
	ErrCurrenciesMismatch = Error("mismatch of currencies")
)
