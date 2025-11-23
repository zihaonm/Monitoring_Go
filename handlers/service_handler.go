package handlers

import (
	"monitoring/models"
	"monitoring/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ServiceHandler handles HTTP requests for service management
type ServiceHandler struct {
	store   *models.ServiceStore
	history *models.HistoryStore
	monitor *services.MonitorService
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(store *models.ServiceStore, history *models.HistoryStore, monitor *services.MonitorService) *ServiceHandler {
	return &ServiceHandler{
		store:   store,
		history: history,
		monitor: monitor,
	}
}

// CreateService handles POST /api/services
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req models.MonitoredService

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	req.ID = uuid.New().String()
	req.Status = models.StatusUnknown
	req.CreatedAt = time.Now()

	if req.CheckInterval == 0 {
		req.CheckInterval = 60 // default to 60 seconds
	}

	if req.Timeout == 0 {
		req.Timeout = 10 // default to 10 seconds
	}

	if err := h.store.Add(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Perform initial health check
	result := h.monitor.CheckService(&req)
	if err := h.monitor.UpdateServiceStatus(result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated service
	service, _ := h.store.Get(req.ID)

	c.JSON(http.StatusCreated, service)
}

// GetAllServices handles GET /api/services
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	services := h.store.GetAll()
	c.JSON(http.StatusOK, services)
}

// GetService handles GET /api/services/:id
func (h *ServiceHandler) GetService(c *gin.Context) {
	id := c.Param("id")

	service, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	c.JSON(http.StatusOK, service)
}

// UpdateService handles PUT /api/services/:id
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id := c.Param("id")

	existing, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	var req models.MonitoredService
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve certain fields
	req.ID = existing.ID
	req.CreatedAt = existing.CreatedAt
	req.Status = existing.Status
	req.LastCheck = existing.LastCheck

	if err := h.store.Update(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, &req)
}

// DeleteService handles DELETE /api/services/:id
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Also delete history for this service
	h.history.DeleteHistory(id)

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted successfully"})
}

// CheckServiceNow handles POST /api/services/:id/check
func (h *ServiceHandler) CheckServiceNow(c *gin.Context) {
	id := c.Param("id")

	service, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	result := h.monitor.CheckService(service)
	if err := h.monitor.UpdateServiceStatus(result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated service
	service, _ = h.store.Get(id)

	c.JSON(http.StatusOK, service)
}

// GetServiceStatistics handles GET /api/services/:id/statistics
func (h *ServiceHandler) GetServiceStatistics(c *gin.Context) {
	id := c.Param("id")

	// Check if service exists
	_, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	stats := h.history.GetStatistics(id)
	c.JSON(http.StatusOK, stats)
}

// GetServiceHistory handles GET /api/services/:id/history
func (h *ServiceHandler) GetServiceHistory(c *gin.Context) {
	id := c.Param("id")

	// Check if service exists
	_, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	history := h.history.GetHistory(id)
	if history == nil {
		c.JSON(http.StatusOK, &models.ServiceHistory{
			ServiceID: id,
			Checks:    []models.HealthCheckRecord{},
		})
		return
	}

	c.JSON(http.StatusOK, history)
}
