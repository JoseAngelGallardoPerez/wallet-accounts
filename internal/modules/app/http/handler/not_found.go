package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFoundHandler
type NotFoundHandler struct{}

// NewCorsHandler creates new cors handler
func NewNotFoundHandler() *NotFoundHandler {
	return &NotFoundHandler{}
}

// NotFoundHandler returns 404 error
func (h NotFoundHandler) NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
}
