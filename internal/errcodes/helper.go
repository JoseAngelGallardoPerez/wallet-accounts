package errcodes

import (
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"
)

func AddError(c *gin.Context, code string) {
	publicErr := &errors.PublicError{
		Code:       code,
		HttpStatus: HttpStatusCodeByErrCode(code),
	}
	errors.AddErrors(c, publicErr)
}

func AddErrorMeta(c *gin.Context, code string, meta interface{}) {
	publicErr := &errors.PublicError{
		Code:       code,
		HttpStatus: HttpStatusCodeByErrCode(code),
		Meta:       meta,
	}
	errors.AddErrors(c, publicErr)
}
