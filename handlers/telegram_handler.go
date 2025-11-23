package handlers

import (
	"monitoring/models"
	"monitoring/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TelegramHandler handles HTTP requests for Telegram configuration
type TelegramHandler struct {
	telegram *services.TelegramService
}

// NewTelegramHandler creates a new Telegram handler
func NewTelegramHandler(telegram *services.TelegramService) *TelegramHandler {
	return &TelegramHandler{
		telegram: telegram,
	}
}

// GetConfig handles GET /api/telegram/config
func (h *TelegramHandler) GetConfig(c *gin.Context) {
	config := h.telegram.GetConfig()
	c.JSON(http.StatusOK, config)
}

// UpdateConfig handles PUT /api/telegram/config
func (h *TelegramHandler) UpdateConfig(c *gin.Context) {
	var req models.TelegramConfig

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.telegram.SetConfig(&req)

	c.JSON(http.StatusOK, gin.H{
		"message": "Telegram configuration updated successfully",
		"config":  h.telegram.GetConfig(),
	})
}

// TestNotification handles POST /api/telegram/test
func (h *TelegramHandler) TestNotification(c *gin.Context) {
	if err := h.telegram.SendTestMessage(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send test message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test notification sent successfully",
	})
}
