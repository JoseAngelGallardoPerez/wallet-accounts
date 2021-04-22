package middleware

import (
	"strconv"

	"github.com/Confialink/wallet-pkg-errors"
	"github.com/gin-gonic/gin"

	"github.com/Confialink/wallet-accounts/internal/modules/request/repository"
)

func RequestedRequest(repo repository.RequestRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId, err := strconv.ParseUint(c.Param("requestId"), 10, 64)
		if err != nil {
			privateError := errors.PrivateError{Message: err.Error()}
			errors.AddErrors(c, &privateError)
			c.Abort()
			return
		}

		req, err := repo.FindById(requestId)
		if err != nil {
			privateError := errors.PrivateError{Message: err.Error()}
			errors.AddErrors(c, &privateError)
			c.Abort()
			return
		}

		c.Set("_requested_request", req)
	}
}
