package transfers

import "github.com/Confialink/wallet-accounts/internal/errcodes"

// Error defines string error
type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

const (
	ErrUnexpectedStatus       = Error("unexpected transfer request status")
	ErrModificationNotAllowed = Error("modification not allowed")
	ErrMissingInputData       = Error("required input data is missing")
	ErrSubjectNotSupported    = Error("subject not supported")
	ErrMissingRequestData     = Error("required request data is missing")

	ErrWithdrawalNotAllowed = Error(errcodes.CodeWithdrawalNotAllowed)
	ErrDepositNotAllowed    = Error(errcodes.CodeDepositNotAllowed)
	ErrInsufficientBalance  = Error(errcodes.CodeInsufficientFunds)
	ErrAccountInactive      = Error(errcodes.CodeAccountInactive)
)
