package handlers

import (
	"monitoring/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SystemHandler handles HTTP requests for system information
type SystemHandler struct {
	system *services.SystemService
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(system *services.SystemService) *SystemHandler {
	return &SystemHandler{
		system: system,
	}
}

// GetSystemInfo handles GET /api/system/info
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.system.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}
