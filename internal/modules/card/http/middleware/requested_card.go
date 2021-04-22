package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/card/repository"
)

func RequestedCard(contextService appHttpService.ContextInterface, repo repository.CardRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, typedErr := contextService.GetIdParam(c)
		if typedErr != nil {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}

		card, err := repo.Get(uint32(id), nil)
		if err != nil {
			errcodes.AddError(c, errcodes.CodeCardNotFound)
			c.Abort()
			return
		}

		c.Set("_requested_card", card)
	}
}
