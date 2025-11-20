package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"waf-admin/internal/config"
	"waf-admin/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type AuditService struct {
	logger *logrus.Logger
}

func NewAuditService(cfg *config.Config, logger *logrus.Logger) *AuditService {
	return &AuditService{
		logger: logger,
	}
}

func (s *AuditService) LogChange(ctx context.Context, auditLog models.AuditLog) error {
	// Generate unique ID if not provided
	if auditLog.ID == "" {
		auditLog.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if auditLog.Timestamp.IsZero() {
		auditLog.Timestamp = time.Now()
	}

	// Convert audit log to JSON for structured logging
	auditJSON, err := json.Marshal(auditLog)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	// Log to stdout with structured format for log collection
	s.logger.WithFields(logrus.Fields{
		"type":       "audit",
		"audit_data": string(auditJSON),
	}).Info("WAF Audit Log")

	return nil
}

func (s *AuditService) GetAuditLogs(ctx context.Context, limit int, offset int) ([]models.AuditLog, int, error) {
	// Audit logs are now output to stdout and collected by log aggregation system
	// This method is kept for API compatibility but returns empty results
	// Logs should be queried from VictoriaLogs or other log aggregation system
	s.logger.Warn("GetAuditLogs called but audit logs are now output to stdout for collection by log aggregation system")
	return []models.AuditLog{}, 0, nil
}

func (s *AuditService) GetAuditLogsByResource(ctx context.Context, resource string, resourceID string) ([]models.AuditLog, error) {
	// Audit logs are now output to stdout and collected by log aggregation system
	// This method is kept for API compatibility but returns empty results
	// Logs should be queried from VictoriaLogs or other log aggregation system
	s.logger.Warn("GetAuditLogsByResource called but audit logs are now output to stdout for collection by log aggregation system")
	return []models.AuditLog{}, nil
}

func (s *AuditService) CreateAuditLog(action, resource, resourceID, user, ip, userAgent string, oldValue, newValue interface{}) models.AuditLog {
	// Calculate diff
	diff := s.calculateDiff(oldValue, newValue)

	return models.AuditLog{
		ID:         uuid.New().String(),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Diff:       diff,
		User:       user,
		Timestamp:  time.Now(),
		IP:         ip,
		UserAgent:  userAgent,
	}
}

func (s *AuditService) calculateDiff(oldValue, newValue interface{}) string {
	if oldValue == nil && newValue == nil {
		return "No changes"
	}

	if oldValue == nil {
		return "Created"
	}

	if newValue == nil {
		return "Deleted"
	}

	// Convert to JSON for comparison
	oldJSON, _ := json.Marshal(oldValue)
	newJSON, _ := json.Marshal(newValue)

	if string(oldJSON) == string(newJSON) {
		return "No changes"
	}

	return fmt.Sprintf("Changed from %s to %s", string(oldJSON), string(newJSON))
}