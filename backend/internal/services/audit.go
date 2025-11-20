package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"waf-admin/internal/config"
	"waf-admin/internal/k8s"
	"waf-admin/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AuditService struct {
	k8sClient *k8s.Client
	config    *config.Config
	logger    *logrus.Logger
}

func NewAuditService(k8sClient *k8s.Client, cfg *config.Config, logger *logrus.Logger) *AuditService {
	return &AuditService{
		k8sClient: k8sClient,
		config:    cfg,
		logger:    logger,
	}
}

func (s *AuditService) LogChange(ctx context.Context, auditLog models.AuditLog) error {
	configMap, err := s.getAuditConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get audit configmap: %w", err)
	}

	// Generate unique ID if not provided
	if auditLog.ID == "" {
		auditLog.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if auditLog.Timestamp.IsZero() {
		auditLog.Timestamp = time.Now()
	}

	// Get existing audit logs
	var auditLogs []models.AuditLog
	if auditData, exists := configMap.Data["audit.log"]; exists && auditData != "" {
		if err := json.Unmarshal([]byte(auditData), &auditLogs); err != nil {
			s.logger.Warnf("Failed to unmarshal audit logs: %v", err)
			auditLogs = []models.AuditLog{}
		}
	}

	// Add new audit log
	auditLogs = append(auditLogs, auditLog)

	// Keep only last 1000 entries to prevent configmap from growing too large
	if len(auditLogs) > 1000 {
		auditLogs = auditLogs[len(auditLogs)-1000:]
	}

	// Marshal back to JSON
	auditData, err := json.Marshal(auditLogs)
	if err != nil {
		return fmt.Errorf("failed to marshal audit logs: %w", err)
	}

	configMap.Data["audit.log"] = string(auditData)

	return s.k8sClient.UpdateConfigMap(ctx, s.config.Kubernetes.Namespace, configMap)
}

func (s *AuditService) GetAuditLogs(ctx context.Context, limit int, offset int) ([]models.AuditLog, int, error) {
	configMap, err := s.getAuditConfigMap(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit configmap: %w", err)
	}

	var auditLogs []models.AuditLog
	if auditData, exists := configMap.Data["audit.log"]; exists && auditData != "" {
		if err := json.Unmarshal([]byte(auditData), &auditLogs); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal audit logs: %w", err)
		}
	}

	total := len(auditLogs)

	// Apply pagination
	start := offset
	end := offset + limit
	if start > total {
		return []models.AuditLog{}, total, nil
	}
	if end > total {
		end = total
	}

	return auditLogs[start:end], total, nil
}

func (s *AuditService) GetAuditLogsByResource(ctx context.Context, resource string, resourceID string) ([]models.AuditLog, error) {
	configMap, err := s.getAuditConfigMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit configmap: %w", err)
	}

	var allLogs []models.AuditLog
	if auditData, exists := configMap.Data["audit.log"]; exists && auditData != "" {
		if err := json.Unmarshal([]byte(auditData), &allLogs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal audit logs: %w", err)
		}
	}

	var filteredLogs []models.AuditLog
	for _, log := range allLogs {
		if log.Resource == resource && (resourceID == "" || log.ResourceID == resourceID) {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return filteredLogs, nil
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

func (s *AuditService) getAuditConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap, err := s.k8sClient.GetConfigMap(ctx, s.config.Kubernetes.Namespace, "waf-audit-logs")
	if err != nil {
		if errors.IsNotFound(err) {
			return s.createAuditConfigMap(ctx)
		}
		return nil, err
	}
	return configMap, nil
}

func (s *AuditService) createAuditConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "waf-audit-logs",
			Namespace: s.config.Kubernetes.Namespace,
			Labels: map[string]string{
				"app": "waf-admin",
			},
		},
		Data: map[string]string{
			"audit.log": "[]",
		},
	}

	return s.k8sClient.CreateConfigMap(ctx, s.config.Kubernetes.Namespace, configMap)
}