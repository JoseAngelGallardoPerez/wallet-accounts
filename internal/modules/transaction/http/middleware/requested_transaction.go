package middleware

import (
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/transaction/repository"
)

func RequestedTransaction(contextService appHttpService.ContextInterface, repo *repository.TransactionRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, typedErr := contextService.GetIdParam(c)
		if typedErr != nil {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}

		transaction, e := repo.GetById(id)
		if e != nil {
			privateError := errors.PrivateError{Message: e.Error()}
			errors.AddErrors(c, &privateError)
			return
		}

		c.Set("_requested_transaction", transaction)
	}
}
