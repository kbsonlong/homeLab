package api

import (
	"net/http"
	"time"

	"waf-admin/internal/models"
	"waf-admin/internal/services"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	metricsService *services.MetricsService
}

func NewMetricsHandler(metricsService *services.MetricsService) *MetricsHandler {
	return &MetricsHandler{
		metricsService: metricsService,
	}
}

// GetMetricsSummary returns aggregated metrics
func (h *MetricsHandler) GetMetricsSummary(c *gin.Context) {
	// Parse time range from query parameters
	startStr := c.DefaultQuery("start", time.Now().Add(-1*time.Hour).Format(time.RFC3339))
	endStr := c.DefaultQuery("end", time.Now().Format(time.RFC3339))

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start time format"})
		return
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end time format"})
		return
	}

	timeRange := models.TimeRange{
		Start: start,
		End:   end,
	}

	summary, err := h.metricsService.GetMetricsSummary(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

type LogsHandler struct {
	logsService *services.LogsService
}

func NewLogsHandler(logsService *services.LogsService) *LogsHandler {
	return &LogsHandler{
		logsService: logsService,
	}
}

// SearchLogs searches logs based on query
func (h *LogsHandler) SearchLogs(c *gin.Context) {
	var query models.LogQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.logsService.SearchLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search logs"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetLogFilters returns available log filters
func (h *LogsHandler) GetLogFilters(c *gin.Context) {
	filters := h.logsService.GetLogFilters()
	c.JSON(http.StatusOK, filters)
}