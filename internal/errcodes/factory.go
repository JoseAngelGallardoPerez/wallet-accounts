package errcodes

import (
	"github.com/Confialink/wallet-pkg-errors"
)

func CreatePublicError(code string, title ...string) *errors.PublicError {
	err := &errors.PublicError{Code: code}
	if details, ok := details[code]; ok {
		err.Details = details
	}

	if len(title) > 0 {
		err.Title = title[0]
	}

	err.HttpStatus = HttpStatusCodeByErrCode(code)

	return err
}
