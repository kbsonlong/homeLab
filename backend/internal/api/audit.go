package api

import (
	"net/http"
	"strconv"

	"waf-admin/internal/models"
	"waf-admin/internal/services"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService *services.AuditService
}

func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs returns audit logs with pagination
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	resource := c.Query("resource")
	resourceID := c.Query("resource_id")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	var logs []interface{}
	var total int
	var auditErr error

	if resource != "" {
		// Filter by resource
		var auditLogs []models.AuditLog
		auditLogs, auditErr = h.auditService.GetAuditLogsByResource(c.Request.Context(), resource, resourceID)
		if auditErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit logs"})
			return
		}
		total = len(auditLogs)
		// Convert to interface slice
		logs = make([]interface{}, len(auditLogs))
		for i, log := range auditLogs {
			logs[i] = log
		}
		// Apply pagination manually for filtered results
		if offset < len(logs) {
			end := offset + limit
			if end > len(logs) {
				end = len(logs)
			}
			logs = logs[offset:end]
		} else {
			logs = []interface{}{}
		}
	} else {
		// Get all logs with pagination
		var auditLogs []models.AuditLog
		auditLogs, total, auditErr = h.auditService.GetAuditLogs(c.Request.Context(), limit, offset)
		if auditErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit logs"})
			return
		}
		logs = make([]interface{}, len(auditLogs))
		for i, log := range auditLogs {
			logs[i] = log
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetAuditLog returns a specific audit log by ID
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	logID := c.Param("id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Log ID is required"})
		return
	}

	// Get all logs and find by ID
	logs, _, err := h.auditService.GetAuditLogs(c.Request.Context(), 1000, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit log"})
		return
	}

	for _, log := range logs {
		if log.ID == logID {
			c.JSON(http.StatusOK, log)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Audit log not found"})
}