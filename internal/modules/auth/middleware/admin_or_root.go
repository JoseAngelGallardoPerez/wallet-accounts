package middleware

import (
	"net/http"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	"github.com/Confialink/wallet-users/rpc/proto/users"
	"github.com/gin-gonic/gin"
)

func AdminOrRoot() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, ok := ctx.Get("_user")
		if !ok {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		roleName := user.(*users.User).RoleName
		if roleName != "admin" && roleName != "root" {
			ctx.Status(http.StatusForbidden)
			_ = ctx.Error(errcodes.CreatePublicError(errcodes.CodeForbidden, "you are not allowed to perform this action"))
			ctx.Abort()
			return
		}
	}
}
