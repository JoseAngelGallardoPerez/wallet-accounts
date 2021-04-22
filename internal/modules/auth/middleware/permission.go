package middleware

import (
	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/errcodes"
	appHttpService "github.com/Confialink/wallet-accounts/internal/modules/app/http/service"
	authService "github.com/Confialink/wallet-accounts/internal/modules/auth/service"
)

type PermissionChecker struct {
	authService    authService.AuthServiceInterface
	contextService appHttpService.ContextInterface
}

func NewPermissionChecker(authService authService.AuthServiceInterface, contextService appHttpService.ContextInterface) *PermissionChecker {
	return &PermissionChecker{authService, contextService}
}

// check dynamic permission for resource
func (s *PermissionChecker) CanDynamic(action string, resourceName string, resource interface{}) func(*gin.Context) {
	return func(c *gin.Context) {
		user := s.contextService.MustGetCurrentUser(c)
		if !s.authService.CanDynamic(user, action, resourceName, resource) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}

// check permission for resource
func (s *PermissionChecker) Can(action string, resource string) func(*gin.Context) {
	return func(c *gin.Context) {
		user := s.contextService.MustGetCurrentUser(c)
		if !s.authService.Can(user.RoleName, action, resource) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}

// check permissions for Account resource. Account is in context
func (s *PermissionChecker) CanDynamicWithAccount(action string, resource string) func(*gin.Context) {
	return func(c *gin.Context) {
		user := s.contextService.MustGetCurrentUser(c)
		account := s.contextService.GetRequestedAccount(c)
		if account == nil {
			errcodes.AddError(c, errcodes.CodeAccountNotFound)
			c.Abort()
			return
		}

		if !s.authService.CanDynamic(user, action, resource, account) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}

// check permissions for Card resource. Card is in context
func (s *PermissionChecker) CanDynamicWithCard(action string, resource string) func(*gin.Context) {
	return func(c *gin.Context) {
		user := s.contextService.MustGetCurrentUser(c)
		card := s.contextService.GetRequestedCard(c)
		if card == nil {
			errcodes.AddError(c, errcodes.CodeCardNotFound)
			c.Abort()
			return
		}

		if !s.authService.CanDynamic(user, action, resource, card) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}

// check permissions for Request resource. Request is in context
func (s *PermissionChecker) CanDynamicWithRequest(action string, resource string) func(*gin.Context) {
	return func(c *gin.Context) {
		user := s.contextService.MustGetCurrentUser(c)
		request := s.contextService.GetRequestedRequest(c)
		if request == nil {
			privateError := errors.PrivateError{Message: "request not found"}
			errors.AddErrors(c, &privateError)
			c.Abort()
			return
		}

		if !s.authService.CanDynamic(user, action, resource, request) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}

// check permissions for Transaction resource. Transaction is in context
func (s *PermissionChecker) CanDynamicWithTransaction(action string, resource string) func(*gin.Context) {
	return func(c *gin.Context) {
		user := s.contextService.MustGetCurrentUser(c)
		transaction := s.contextService.GetRequestedTransaction(c)
		if transaction == nil {
			privateError := errors.PrivateError{Message: "transaction not found"}
			errors.AddErrors(c, &privateError)
			c.Abort()
			return
		}

		if !s.authService.CanDynamic(user, action, resource, transaction) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}

// check permissions for Param resource. Param is in context
func (s *PermissionChecker) CanDynamicWithParam(action string, resource string, paramName string) func(*gin.Context) {
	return func(c *gin.Context) {
		param := c.Param(paramName)
		user := s.contextService.MustGetCurrentUser(c)
		if !s.authService.CanDynamic(user, action, resource, param) {
			errcodes.AddError(c, errcodes.CodeForbidden)
			c.Abort()
			return
		}
	}
}
