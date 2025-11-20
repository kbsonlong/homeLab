package services

import (
	"context"
	"fmt"
	"time"

	"waf-admin/internal/config"
	"waf-admin/internal/k8s"
	"waf-admin/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type WAFService struct {
	k8sClient    *k8s.Client
	config       *config.Config
	logger       *logrus.Logger
	auditService *AuditService
}

func NewWAFService(k8sClient *k8s.Client, cfg *config.Config, logger *logrus.Logger) *WAFService {
	return &WAFService{
		k8sClient: k8sClient,
		config:    cfg,
		logger:    logger,
	}
}

func (s *WAFService) SetAuditService(auditService *AuditService) {
	s.auditService = auditService
}

func (s *WAFService) GetWAFStatus(ctx context.Context) (*models.WAFStatus, error) {
	configMap, err := s.k8sClient.GetWAFPolicyConfigMap(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get WAF policy configmap: %w", err)
	}

	status := &models.WAFStatus{
		HostPolicies: make(map[string]models.WAFPolicy),
		LastUpdated:  time.Now(),
	}

	if policiesData, exists := configMap.Data["policies.yaml"]; exists && policiesData != "{}" {
		var policies map[string]models.WAFPolicy
		if err := yaml.Unmarshal([]byte(policiesData), &policies); err != nil {
			s.logger.Warnf("Failed to unmarshal policies: %v", err)
		} else {
			status.HostPolicies = policies
			if globalPolicy, exists := policies["global"]; exists {
				status.GlobalPolicy = globalPolicy
			}
		}
	}

	controllerConfigMap, err := s.k8sClient.GetIngressNGINXControllerConfigMap(ctx)
	if err == nil {
		status.ControllerConfig = models.ControllerConfig{
			AllowSnippetAnnotations: controllerConfigMap.Data["allow-snippet-annotations"] == "true",
			ModSecuritySnippet:      controllerConfigMap.Data["modsecurity-snippet"],
		}
	}

	return status, nil
}

func (s *WAFService) UpdateWAFMode(ctx context.Context, req models.PolicyUpdateRequest) error {
	configMap, err := s.k8sClient.GetWAFPolicyConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WAF policy configmap: %w", err)
	}

	var policies map[string]models.WAFPolicy
	if policiesData, exists := configMap.Data["policies.yaml"]; exists && policiesData != "{}" {
		if err := yaml.Unmarshal([]byte(policiesData), &policies); err != nil {
			return fmt.Errorf("failed to unmarshal policies: %w", err)
		}
	} else {
		policies = make(map[string]models.WAFPolicy)
	}

	policy, exists := policies[req.Host]
	if !exists {
		policy = models.WAFPolicy{
			ID:        uuid.New().String(),
			Host:      req.Host,
			CreatedAt: time.Now(),
		}
	}

	policy.Mode = req.Mode
	policy.UpdatedAt = time.Now()
	policy.Version++

	if req.EnableCRS != nil {
		policy.EnableCRS = *req.EnableCRS
	}

	policies[req.Host] = policy

	policiesData, err := yaml.Marshal(policies)
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	configMap.Data["policies.yaml"] = string(policiesData)

	if err := s.k8sClient.UpdateConfigMap(ctx, s.config.Kubernetes.Namespace, configMap); err != nil {
		return fmt.Errorf("failed to update configmap: %w", err)
	}

	// Log the change
	if s.auditService != nil {
		auditLog := s.auditService.CreateAuditLog(
			"UPDATE_MODE",
			"waf_policy",
			req.Host,
			"system",
			"",
			"",
			policy,
			policy,
		)
		if err := s.auditService.LogChange(ctx, auditLog); err != nil {
			s.logger.Warnf("Failed to log audit change: %v", err)
		}
	}

	return nil
}

func (s *WAFService) UpdateExceptions(ctx context.Context, req models.ExceptionUpdateRequest) error {
	configMap, err := s.k8sClient.GetWAFPolicyConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WAF policy configmap: %w", err)
	}

	var policies map[string]models.WAFPolicy
	if policiesData, exists := configMap.Data["policies.yaml"]; exists && policiesData != "{}" {
		if err := yaml.Unmarshal([]byte(policiesData), &policies); err != nil {
			return fmt.Errorf("failed to unmarshal policies: %w", err)
		}
	} else {
		policies = make(map[string]models.WAFPolicy)
	}

	policy, exists := policies[req.Host]
	if !exists {
		policy = models.WAFPolicy{
			ID:        uuid.New().String(),
			Host:      req.Host,
			CreatedAt: time.Now(),
		}
	}

	policy.Exceptions = req.Exceptions
	policy.UpdatedAt = time.Now()
	policy.Version++

	policies[req.Host] = policy

	policiesData, err := yaml.Marshal(policies)
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	configMap.Data["policies.yaml"] = string(policiesData)

	if err := s.k8sClient.UpdateConfigMap(ctx, s.config.Kubernetes.Namespace, configMap); err != nil {
		return fmt.Errorf("failed to update configmap: %w", err)
	}

	if !req.TestMode {
		if err := s.applyPolicy(ctx, req.Host, policy); err != nil {
			return err
		}
	}

	// Log the change
	if s.auditService != nil {
		auditLog := s.auditService.CreateAuditLog(
			"UPDATE_EXCEPTIONS",
			"waf_policy",
			req.Host,
			"system",
			"",
			"",
			policy,
			policy,
		)
		if err := s.auditService.LogChange(ctx, auditLog); err != nil {
			s.logger.Warnf("Failed to log audit change: %v", err)
		}
	}

	return nil
}

func (s *WAFService) UpdateRules(ctx context.Context, req models.RuleUpdateRequest) error {
	configMap, err := s.k8sClient.GetWAFPolicyConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WAF policy configmap: %w", err)
	}

	var policies map[string]models.WAFPolicy
	if policiesData, exists := configMap.Data["policies.yaml"]; exists && policiesData != "{}" {
		if err := yaml.Unmarshal([]byte(policiesData), &policies); err != nil {
			return fmt.Errorf("failed to unmarshal policies: %w", err)
		}
	} else {
		policies = make(map[string]models.WAFPolicy)
	}

	policy, exists := policies[req.Host]
	if !exists {
		policy = models.WAFPolicy{
			ID:        uuid.New().String(),
			Host:      req.Host,
			CreatedAt: time.Now(),
		}
	}

	policy.CustomRules = req.CustomRules
	policy.UpdatedAt = time.Now()
	policy.Version++

	policies[req.Host] = policy

	policiesData, err := yaml.Marshal(policies)
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	configMap.Data["policies.yaml"] = string(policiesData)

	if err := s.k8sClient.UpdateConfigMap(ctx, s.config.Kubernetes.Namespace, configMap); err != nil {
		return fmt.Errorf("failed to update configmap: %w", err)
	}

	if err := s.applyPolicy(ctx, req.Host, policy); err != nil {
		return err
	}

	// Log the change
	if s.auditService != nil {
		auditLog := s.auditService.CreateAuditLog(
			"UPDATE_RULES",
			"waf_policy",
			req.Host,
			"system",
			"",
			"",
			policy,
			policy,
		)
		if err := s.auditService.LogChange(ctx, auditLog); err != nil {
			s.logger.Warnf("Failed to log audit change: %v", err)
		}
	}

	return nil
}

func (s *WAFService) ApplyConfiguration(ctx context.Context, req models.ApplyRequest) error {
	configMap, err := s.k8sClient.GetWAFPolicyConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WAF policy configmap: %w", err)
	}

	var policies map[string]models.WAFPolicy
	if policiesData, exists := configMap.Data["policies.yaml"]; exists && policiesData != "{}" {
		if err := yaml.Unmarshal([]byte(policiesData), &policies); err != nil {
			return fmt.Errorf("failed to unmarshal policies: %w", err)
		}
	} else {
		return fmt.Errorf("no policies found for host: %s", req.Host)
	}

	policy, exists := policies[req.Host]
	if !exists {
		return fmt.Errorf("no policy found for host: %s", req.Host)
	}

	if req.Strategy == "annotation" {
		if err := s.k8sClient.ApplyWAFPolicyToIngress(ctx, req.Host, policy); err != nil {
			return fmt.Errorf("failed to apply policy to ingress: %w", err)
		}
	} else {
		if err := s.k8sClient.ApplyWAFPolicyToController(ctx, policy); err != nil {
			return fmt.Errorf("failed to apply policy to controller: %w", err)
		}
	}

	if err := s.k8sClient.RolloutDeployment(ctx, "ingress-nginx", "ingress-nginx-controller"); err != nil {
		return err
	}

	// Log the change
	if s.auditService != nil {
		auditLog := s.auditService.CreateAuditLog(
			"APPLY_CONFIGURATION",
			"waf_policy",
			req.Host,
			"system",
			req.Strategy,
			"",
			policy,
			policy,
		)
		if err := s.auditService.LogChange(ctx, auditLog); err != nil {
			s.logger.Warnf("Failed to log audit change: %v", err)
		}
	}

	return nil
}

func (s *WAFService) applyPolicy(ctx context.Context, host string, policy models.WAFPolicy) error {
	if s.config.Kubernetes.Namespace == "ingress-nginx" {
		return s.k8sClient.ApplyWAFPolicyToController(ctx, policy)
	}
	return s.k8sClient.ApplyWAFPolicyToIngress(ctx, host, policy)
}