package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"waf-admin/internal/config"
	"waf-admin/internal/models"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *config.Config
	logger    *logrus.Logger
}

func NewClient(cfg *config.Config, logger *logrus.Logger) (*Client, error) {
	var kubeConfig *rest.Config
	var err error

	if cfg.Kubernetes.ConfigPath != "" {
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubernetes.ConfigPath)
	} else {
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
			if _, err := os.Stat(kubeconfig); err == nil {
				kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    cfg,
		logger:    logger,
	}, nil
}

func (c *Client) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) CreateConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
}

func (c *Client) UpdateConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) error {
	_, err := c.clientset.CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	return err
}

func (c *Client) GetIngress(ctx context.Context, namespace, name string) (*networkingv1.Ingress, error) {
	return c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) UpdateIngress(ctx context.Context, namespace string, ingress *networkingv1.Ingress) error {
	_, err := c.clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	return err
}

func (c *Client) PatchIngress(ctx context.Context, namespace, name string, patch []byte) error {
	_, err := c.clientset.NetworkingV1().Ingresses(namespace).Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) RolloutDeployment(ctx context.Context, namespace, name string) error {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	return err
}

func (c *Client) GetWAFPolicyConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap, err := c.GetConfigMap(ctx, c.config.Kubernetes.Namespace, "waf-policies")
	if err != nil {
		if errors.IsNotFound(err) {
			return c.createWAFPolicyConfigMap(ctx)
		}
		return nil, err
	}
	return configMap, nil
}

func (c *Client) createWAFPolicyConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "waf-policies",
			Namespace: c.config.Kubernetes.Namespace,
			Labels: map[string]string{
				"app": "waf-admin",
			},
		},
		Data: map[string]string{
			"policies.yaml": "{}",
		},
	}

	return c.clientset.CoreV1().ConfigMaps(c.config.Kubernetes.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
}

func (c *Client) GetIngressNGINXControllerConfigMap(ctx context.Context) (*corev1.ConfigMap, error) {
	return c.GetConfigMap(ctx, "ingress-nginx", "ingress-nginx-controller")
}

func (c *Client) ApplyWAFPolicyToIngress(ctx context.Context, host string, policy models.WAFPolicy) error {
	ingressList, err := c.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list ingresses: %w", err)
	}

	// Try exact match first
	for _, ingress := range ingressList.Items {
		for _, rule := range ingress.Spec.Rules {
			if rule.Host == host {
				return c.applyPolicyToIngress(ctx, &ingress, policy)
			}
		}
	}
	
	// If no exact match found, try to create a new ingress for the host
	return c.createIngressForHost(ctx, host, policy)
}

func (c *Client) applyPolicyToIngress(ctx context.Context, ingress *networkingv1.Ingress, policy models.WAFPolicy) error {
	if ingress.Annotations == nil {
		ingress.Annotations = make(map[string]string)
	}

	if policy.Mode == string(models.WAFModeOn) {
		ingress.Annotations["nginx.ingress.kubernetes.io/enable-modsecurity"] = "true"
		ingress.Annotations["nginx.ingress.kubernetes.io/enable-owasp-core-rules"] = "true"
		
		// Always generate snippet to include exceptions and custom rules
		snippet := c.generateModSecuritySnippet(policy)
		if snippet != "" {
			ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = snippet
		} else {
			delete(ingress.Annotations, "nginx.ingress.kubernetes.io/modsecurity-snippet")
		}
	} else if policy.Mode == string(models.WAFModeDetectionOnly) {
		ingress.Annotations["nginx.ingress.kubernetes.io/enable-modsecurity"] = "true"
		ingress.Annotations["nginx.ingress.kubernetes.io/enable-owasp-core-rules"] = "true"
		
		// Generate snippet for detection mode with exceptions
		snippet := "SecRuleEngine DetectionOnly\n"
		exceptionSnippet := c.generateModSecuritySnippet(policy)
		if exceptionSnippet != "" {
			snippet += exceptionSnippet
		}
		ingress.Annotations["nginx.ingress.kubernetes.io/modsecurity-snippet"] = snippet
	} else {
		delete(ingress.Annotations, "nginx.ingress.kubernetes.io/enable-modsecurity")
		delete(ingress.Annotations, "nginx.ingress.kubernetes.io/enable-owasp-core-rules")
		delete(ingress.Annotations, "nginx.ingress.kubernetes.io/modsecurity-snippet")
	}

	return c.UpdateIngress(ctx, ingress.Namespace, ingress)
}

func (c *Client) createIngressForHost(ctx context.Context, host string, policy models.WAFPolicy) error {
	// Check if a service exists that we can use as backend
	services, err := c.clientset.CoreV1().Services("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}
	
	// Find a suitable backend service
	backendService := "echo-server" // Default to echo-server if available
	backendPort := int32(80)
	
	for _, svc := range services.Items {
		if svc.Name == "echo-server" {
			backendService = "echo-server"
			if len(svc.Spec.Ports) > 0 {
				backendPort = svc.Spec.Ports[0].Port
			}
			break
		} else if svc.Name == "ingress-nginx-defaultbackend" {
			backendService = "ingress-nginx-defaultbackend"
			if len(svc.Spec.Ports) > 0 {
				backendPort = svc.Spec.Ports[0].Port
			}
			break
		}
	}
	
	// Create a new ingress for the host
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("waf-%s", strings.ReplaceAll(host, ".", "-")),
			Namespace: "default",
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
				"description": fmt.Sprintf("Auto-generated by WAF for host %s", host),
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: func() *networkingv1.PathType {
										pt := networkingv1.PathTypePrefix
										return &pt
									}(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: backendService,
											Port: networkingv1.ServiceBackendPort{
												Number: backendPort,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Apply WAF policy to the new ingress
	if err := c.applyPolicyToIngress(ctx, ingress, policy); err != nil {
		return fmt.Errorf("failed to apply policy to new ingress: %w", err)
	}
	
	// Create the ingress
	_, err = c.clientset.NetworkingV1().Ingresses(ingress.Namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ingress for host %s: %w", host, err)
	}
	
	logrus.Infof("Successfully created ingress %s for host %s with WAF policy", ingress.Name, host)
	return nil
}

func (c *Client) generateModSecuritySnippet(policy models.WAFPolicy) string {
	snippet := ""
	
	// Add custom rules
	for _, rule := range policy.CustomRules {
		if rule.Enabled {
			snippet += rule.Rule + "\n"
		}
	}
	
	// Add exception rules
	// Path exceptions
	for _, path := range policy.Exceptions.Paths {
		if path != "" {
			snippet += fmt.Sprintf("SecRule REQUEST_URI \"@streq %s\" \"id:%d,phase:1,nolog,pass,ctl:ruleEngine=Off\"\n", 
				path, 10000+len(snippet))
		}
	}
	
	// Method exceptions  
	for _, method := range policy.Exceptions.Methods {
		if method != "" {
			snippet += fmt.Sprintf("SecRule REQUEST_METHOD \"@streq %s\" \"id:%d,phase:1,nolog,pass,ctl:ruleEngine=Off\"\n", 
				method, 10000+len(snippet))
		}
	}
	
	// IP allowlist exceptions
	if len(policy.Exceptions.IPAllow) > 0 {
		ipList := strings.Join(policy.Exceptions.IPAllow, ",")
		snippet += fmt.Sprintf("SecRule REMOTE_ADDR \"@ipMatch %s\" \"id:%d,phase:1,nolog,pass,ctl:ruleEngine=Off\"\n", 
			ipList, 10000+len(snippet))
	}
	
	return snippet
}

func (c *Client) ApplyWAFPolicyToController(ctx context.Context, policy models.WAFPolicy) error {
	configMap, err := c.GetIngressNGINXControllerConfigMap(ctx)
	if err != nil {
		return fmt.Errorf("failed to get controller configmap: %w", err)
	}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	snippet := c.generateControllerModSecuritySnippet(policy)
	configMap.Data["modsecurity-snippet"] = snippet

	return c.UpdateConfigMap(ctx, "ingress-nginx", configMap)
}

func (c *Client) generateControllerModSecuritySnippet(policy models.WAFPolicy) string {
	snippet := ""
	
	if policy.Mode == string(models.WAFModeOn) {
		snippet += "SecRuleEngine On\n"
	} else if policy.Mode == string(models.WAFModeDetectionOnly) {
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