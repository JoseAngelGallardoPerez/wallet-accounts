package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/account/repository"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
)

func RequestedAccount(contextService appHttpService.ContextInterface, repo *repository.AccountRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, typedErr := contextService.GetIdParam(c)
		if typedErr != nil {
			errcodes.AddError(c, errcodes.CodeForbidden)
			return
		}

		account, err := repo.FindByID(id)
		if err != nil {
			errcodes.AddError(c, errcodes.CodeAccountNotFound)
			c.Abort()
			return
		}
		c.Set("_requested_account", account)
	}
}
