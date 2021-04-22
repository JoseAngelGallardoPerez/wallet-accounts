package handler

import (
	"github.com/gin-gonic/gin"
)

// CorsHandler
type CorsHandler struct{}

// NewCorsHandler creates new cors handler
func NewCorsHandler() *CorsHandler {
	return &CorsHandler{}
}

// OptionsHandler handles option requests
func (h *CorsHandler) OptionsHandler(c *gin.Context) {}
