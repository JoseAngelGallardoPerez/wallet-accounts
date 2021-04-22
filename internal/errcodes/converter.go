package errcodes

import (
	errorsPkg "github.com/Confialink/wallet-pkg-errors"
	"github.com/pkg/errors"
)

// convert a simple error to a typed error.
func ConvertToTyped(err error) errorsPkg.TypedError {
	if knowErr := errors.Cause(err); IsKnownCode(knowErr.Error()) {
		if logger != nil {
			logger.Error("shadowed error", "err", err)
		}
		return CreatePublicError(knowErr.Error())
	}
	if typedErr, ok := err.(errorsPkg.TypedError); ok {
		return typedErr
	}
	return errorsPkg.ShouldBindToTyped(err)
}
