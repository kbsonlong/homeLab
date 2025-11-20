package api

import (
	"net/http"

	"waf-admin/internal/models"
	"waf-admin/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type WAFHandler struct {
	wafService *services.WAFService
	logger     *logrus.Logger
}

func NewWAFHandler(wafService *services.WAFService, logger *logrus.Logger) *WAFHandler {
	return &WAFHandler{
		wafService: wafService,
		logger:     logger,
	}
}

// GetWAFStatus returns the current WAF status
func (h *WAFHandler) GetWAFStatus(c *gin.Context) {
	status, err := h.wafService.GetWAFStatus(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get WAF status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get WAF status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// UpdateWAFMode updates the WAF mode for a specific host
func (h *WAFHandler) UpdateWAFMode(c *gin.Context) {
	var req models.PolicyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.wafService.UpdateWAFMode(c.Request.Context(), req); err != nil {
		h.logger.Errorf("Failed to update WAF mode: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update WAF mode"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "WAF mode updated successfully"})
}

// UpdateExceptions updates exception rules for a specific host
func (h *WAFHandler) UpdateExceptions(c *gin.Context) {
	var req models.ExceptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.wafService.UpdateExceptions(c.Request.Context(), req); err != nil {
		h.logger.Errorf("Failed to update exceptions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exceptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Exceptions updated successfully"})
}

// UpdateRules updates custom rules for a specific host
func (h *WAFHandler) UpdateRules(c *gin.Context) {
	var req models.RuleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.wafService.UpdateRules(c.Request.Context(), req); err != nil {
		h.logger.Errorf("Failed to update rules: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rules updated successfully"})
}

// ApplyConfiguration applies the WAF configuration
func (h *WAFHandler) ApplyConfiguration(c *gin.Context) {
	var req models.ApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.wafService.ApplyConfiguration(c.Request.Context(), req); err != nil {
		h.logger.Errorf("Failed to apply configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration applied successfully"})
}