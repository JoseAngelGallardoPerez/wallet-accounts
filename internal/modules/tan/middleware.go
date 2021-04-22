package tan

import (
	"github.com/Confialink/wallet-accounts/internal/errcodes"

	"github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	"github.com/Confialink/wallet-accounts/internal/modules/settings"
	"github.com/gin-gonic/gin"
)

func MiddlewareUseIfRequired(
	tanService *Service,
	contextService service.ContextInterface,
	service *settings.Service,
	settingName settings.Name,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		isRequired, err := service.Bool(settingName)
		if err != nil {
			panic(err)
		}

		if isRequired {
			tan := retrieveTanHeader(c)
			if tan == "" {
				errcodes.AddError(c, errcodes.CodeTanEmpty)
				c.Abort()
				return
			}

			user := contextService.MustGetCurrentUser(c)
			if !tanService.Use(user.UID, tan) {
				errcodes.AddError(c, errcodes.CodeTanInvalid)
				c.Abort()
				return
			}
		}
	}
}

func retrieveTanHeader(c *gin.Context) string {
	return c.Request.Header.Get("X-TAN")
}
