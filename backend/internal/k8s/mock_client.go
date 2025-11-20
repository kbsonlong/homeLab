package k8s

import (
	"context"
	"fmt"
	"sync"

	"waf-admin/internal/config"
	"waf-admin/internal/models"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/sirupsen/logrus"
)

// MockClient is a mock implementation of Kubernetes client for testing
type MockClient struct {
	config       *config.Config
	logger       *logrus.Logger
	configMaps   map[string]*corev1.ConfigMap
	ingresses    map[string]*networkingv1.Ingress
	deployments  map[string]*mockDeployment
	mutex        sync.RWMutex
}

type mockDeployment struct {
	name      string
	namespace string
	replicas  int32
}

// NewMockClient creates a new mock Kubernetes client
func NewMockClient(cfg *config.Config, logger *logrus.Logger) *MockClient {
	client := &MockClient{
		config:      cfg,
		logger:      logger,
		configMaps:  make(map[string]*corev1.ConfigMap),
		ingresses:   make(map[string]*networkingv1.Ingress),
		deployments: make(map[string]*mockDeployment),
	}
	
	// Initialize with mock data
	client.initializeMockData()
	return client
}

func (c *MockClient) initializeMockData() {
	// Create mock WAF policy ConfigMap
	wafPolicyCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "waf-policies",
			Namespace: c.config.Kubernetes.Namespace,
		},
		Data: map[string]string{
			"policies.yaml": "{}",
		},
	}
	c.configMaps["waf-policies"] = wafPolicyCM
	
	// Create mock ingress-nginx controller ConfigMap
	controllerCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress-nginx-controller",
			Namespace: "ingress-nginx",
		},
		Data: map[string]string{
			"allow-snippet-annotations": "true",
			"modsecurity-snippet":       "",
		},
	}
	c.configMaps["ingress-nginx-controller"] = controllerCM
	
	// Create mock audit ConfigMap
	auditCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "waf-audit-logs",
			Namespace: c.config.Kubernetes.Namespace,
		},
		Data: map[string]string{
			"audit.log": "[]",
		},
	}
	c.configMaps["waf-audit-logs"] = auditCM
}

func (c *MockClient) GetWAFPolicyConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	cm, exists := c.configMaps["waf-policies"]
	if !exists {
		return nil, fmt.Errorf("configmap waf-policies not found")
	}
	return cm, nil
}

func (c *MockClient) GetIngressNGINXControllerConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	cm, exists := c.configMaps["ingress-nginx-controller"]
	if !exists {
		return nil, fmt.Errorf("configmap ingress-nginx-controller not found")
	}
	return cm, nil
}

func (c *MockClient) UpdateConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.configMaps[configMap.Name] = configMap
	c.logger.Infof("Updated ConfigMap %s in namespace %s", configMap.Name, namespace)
	return nil
}

func (c *MockClient) CreateConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.configMaps[configMap.Name] = configMap
	c.logger.Infof("Created ConfigMap %s in namespace %s", configMap.Name, namespace)
	return nil
}

func (c *MockClient) GetIngress(ctx context.Context, name, namespace string) (*networkingv1.Ingress, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	ingress, exists := c.ingresses[name]
	if !exists {
		return nil, fmt.Errorf("ingress %s not found in namespace %s", name, namespace)
	}
	return ingress, nil
}

func (c *MockClient) UpdateIngress(ctx context.Context, namespace string, ingress *networkingv1.Ingress) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.ingresses[ingress.Name] = ingress
	c.logger.Infof("Updated Ingress %s in namespace %s", ingress.Name, namespace)
	return nil
}

func (c *MockClient) ApplyWAFPolicyToIngress(ctx context.Context, host string, policy models.WAFPolicy) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Find ingress by host
	for _, ingress := range c.ingresses {
		for _, rule := range ingress.Spec.Rules {
			if rule.Host == host {
				if ingress.Annotations == nil {
					ingress.Annotations = make(map[string]string)
				}
				
				// Apply WAF annotations based on policy
				if policy.Mode == "On" {
					ingress.Annotations["nginx.ingress.kubernetes.io/enable-modsecurity"] = "true"
					ingress.Annotations["nginx.ingress.kubernetes.io/enable-owasp-core-rules"] = "true"
				} else if policy.Mode == "DetectionOnly" {
					ingress.Annotations["nginx.ingress.kubernetes.io/enable-modsecurity"] = "true"
					ingress.Annotations["nginx.ingress.kubernetes.io/enable-owasp-core-rules"] = "true"
					ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = "SecRuleEngine DetectionOnly"
				} else {
					delete(ingress.Annotations, "nginx.ingress.kubernetes.io/enable-modsecurity")
					delete(ingress.Annotations, "nginx.ingress.kubernetes.io/enable-owasp-core-rules")
					delete(ingress.Annotations, "nginx.ingress.kubernetes.io/modsecurity-snippet")
				}
				
				c.logger.Infof("Applied WAF policy to ingress %s for host %s", ingress.Name, host)
				return nil
			}
		}
	}
	
	return fmt.Errorf("no ingress found for host %s", host)
}

func (c *MockClient) ApplyWAFPolicyToController(ctx context.Context, policy models.WAFPolicy) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	controllerCM, exists := c.configMaps["ingress-nginx-controller"]
	if !exists {
		return fmt.Errorf("controller configmap not found")
	}
	
	if controllerCM.Data == nil {
		controllerCM.Data = make(map[string]string)
	}
	
	// Generate ModSecurity snippet based on policy
	snippet := c.generateModSecuritySnippet(policy)
	controllerCM.Data["modsecurity-snippet"] = snippet
	
	c.logger.Infof("Applied WAF policy to ingress-nginx controller")
	return nil
}

func (c *MockClient) generateModSecuritySnippet(policy models.WAFPolicy) string {
	snippet := ""
	
	if policy.Mode == "On" {
		snippet += "SecRuleEngine On\n"
	} else if policy.Mode == "DetectionOnly" {
		snippet += "SecRuleEngine DetectionOnly\n"
	} else {
		snippet += "SecRuleEngine Off\n"
	}
	
	if policy.EnableCRS {
		snippet += "Include /etc/nginx/modsecurity/modsecurity.conf\n"
		snippet += "Include /etc/nginx/modsecurity/owasp-crs.conf\n"
	}
	
	for _, rule := range policy.CustomRules {
		if rule.Enabled {
			snippet += rule.Rule + "\n"
		}
	}
	
	return snippet
}

func (c *MockClient) RolloutDeployment(ctx context.Context, namespace, deploymentName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.logger.Infof("Mock rollout deployment %s in namespace %s", deploymentName, namespace)
	return nil
}