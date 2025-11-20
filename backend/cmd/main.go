package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"waf-admin/internal/api"
	"waf-admin/internal/config"
	"waf-admin/internal/k8s"
	"waf-admin/internal/models"
	"waf-admin/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Initialize Kubernetes client
	k8sClient, err := k8s.NewClient(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Initialize services
	wafService := services.NewWAFService(k8sClient, cfg, logger)
	auditService := services.NewAuditService(cfg, logger)
	metricsService := services.NewMetricsService(cfg, logger)
	logsService := services.NewLogsService(cfg, logger)

	// Set audit service for WAF service
	wafService.SetAuditService(auditService)

	// Initialize handlers
	wafHandler := api.NewWAFHandler(wafService, logger)
	auditHandler := api.NewAuditHandler(auditService)

	// Setup Gin router
	router := setupRouter(cfg, wafHandler, auditHandler, metricsService, logsService, logger)

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Infof("WAF Admin server started on %s:%d", cfg.Server.Host, cfg.Server.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func setupRouter(cfg *config.Config, wafHandler *api.WAFHandler, auditHandler *api.AuditHandler, metricsService *services.MetricsService, logsService *services.LogsService, logger *logrus.Logger) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Basic auth middleware if enabled
	if cfg.Security.EnableAuth {
		router.Use(gin.BasicAuth(gin.Accounts{
			cfg.Security.Username: cfg.Security.Password,
		}))
	}

	// API routes
	api := router.Group("/api")
	{
		// WAF management
		waf := api.Group("/waf")
		{
			waf.GET("/status", wafHandler.GetWAFStatus)
			waf.POST("/mode", wafHandler.UpdateWAFMode)
			waf.POST("/exceptions", wafHandler.UpdateExceptions)
			waf.POST("/rules", wafHandler.UpdateRules)
			waf.POST("/apply", wafHandler.ApplyConfiguration)
		}

		// Metrics
		metrics := api.Group("/metrics")
		{
			metrics.GET("/summary", func(c *gin.Context) {
				handleMetricsSummary(c, metricsService)
			})
		}

		// Logs
		logs := api.Group("/logs")
		{
			logs.POST("/search", func(c *gin.Context) {
				handleLogsSearch(c, logsService)
			})
			logs.GET("/filters", func(c *gin.Context) {
				c.JSON(http.StatusOK, logsService.GetLogFilters())
			})
		}

		// Audit
		audit := api.Group("/audit")
		{
			audit.GET("", auditHandler.GetAuditLogs)
			audit.GET("/:id", auditHandler.GetAuditLog)
		}

		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})
	}

	return router
}

func handleMetricsSummary(c *gin.Context, service *services.MetricsService) {
	var timeRange models.TimeRange
	if err := c.ShouldBindQuery(&timeRange); err != nil {
		// Default to last 1 hour
		timeRange = models.TimeRange{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		}
	}

	summary, err := service.GetMetricsSummary(c.Request.Context(), timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get metrics summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func handleLogsSearch(c *gin.Context, service *services.LogsService) {
	var query models.LogQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := service.SearchLogs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search logs"})
		return
	}

	c.JSON(http.StatusOK, result)
}