package middleware

import (
	"net/http"

	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-accounts/internal/modules/auth/service"
)

var validRoles = map[string]bool{
	service.RoleClient: true,
}

func Client() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, ok := ctx.Get("_user")
		if !ok {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		roleName := user.(*users.User).RoleName
		if !validRoles[roleName] {
			ctx.Status(http.StatusForbidden)
			_ = ctx.Error(errcodes.CreatePublicError(errcodes.CodeForbidden, "you are not allowed to perform this action"))
			ctx.Abort()
			return
		}
	}
}
