package models

import (
	"time"
)

// WAFPolicy represents a WAF policy for a specific host or globally
type WAFPolicy struct {
    ID          string                 `json:"id" yaml:"id"`
    Host        string                 `json:"host" yaml:"host"`
    Namespace   string                 `json:"namespace" yaml:"namespace"`
    Mode        string                 `json:"mode" yaml:"mode"` // On, DetectionOnly, Off
    EnableCRS   bool                   `json:"enable_crs" yaml:"enable_crs"`
    Exceptions  WAFExceptions          `json:"exceptions" yaml:"exceptions"`
    CustomRules []CustomRule           `json:"custom_rules" yaml:"custom_rules"`
    CreatedAt   time.Time              `json:"created_at" yaml:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at" yaml:"updated_at"`
    UpdatedBy   string                 `json:"updated_by" yaml:"updated_by"`
    Version     int                    `json:"version" yaml:"version"`
}

// WAFExceptions defines exception rules for WAF
type WAFExceptions struct {
	Paths        []string          `json:"paths" yaml:"paths"`
	Methods      []string          `json:"methods" yaml:"methods"`
	IPAllow      []string          `json:"ip_allow" yaml:"ip_allow"`
	HeadersAllow map[string]string `json:"headers_allow" yaml:"headers_allow"`
}

// CustomRule represents a custom ModSecurity rule
type CustomRule struct {
	ID          string    `json:"id" yaml:"id"`
	Name        string    `json:"name" yaml:"name"`
	Rule        string    `json:"rule" yaml:"rule"`
	Description string    `json:"description" yaml:"description"`
	Enabled     bool      `json:"enabled" yaml:"enabled"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
}

// WAFStatus represents the current WAF status
type WAFStatus struct {
	GlobalPolicy WAFPolicy            `json:"global_policy"`
	HostPolicies map[string]WAFPolicy `json:"host_policies"`
	ControllerConfig ControllerConfig   `json:"controller_config"`
	LastUpdated  time.Time            `json:"last_updated"`
}

// ControllerConfig represents ingress-nginx controller configuration
type ControllerConfig struct {
	AllowSnippetAnnotations bool   `json:"allow_snippet_annotations"`
	ModSecuritySnippet    string `json:"modsecurity_snippet"`
}

// MetricsSummary represents aggregated metrics
type MetricsSummary struct {
	TotalRequests   int64              `json:"total_requests"`
	Status4xx       int64              `json:"status_4xx"`
	Status5xx       int64              `json:"status_5xx"`
	Status403       int64              `json:"status_403"`
	WAFBlocked      int64              `json:"waf_blocked"`
	TopHosts        []HostMetrics      `json:"top_hosts"`
	TopPaths        []PathMetrics      `json:"top_paths"`
	TopRuleIDs      []RuleMetrics      `json:"top_rule_ids"`
	TimeRange       TimeRange          `json:"time_range"`
}

// HostMetrics represents metrics for a specific host
type HostMetrics struct {
	Host      string `json:"host"`
	Requests  int64  `json:"requests"`
	Blocked   int64  `json:"blocked"`
	ErrorRate float64 `json:"error_rate"`
}

// PathMetrics represents metrics for a specific path
type PathMetrics struct {
	Path      string `json:"path"`
	Requests  int64  `json:"requests"`
	Blocked   int64  `json:"blocked"`
	ErrorRate float64 `json:"error_rate"`
}

// RuleMetrics represents metrics for a specific rule
type RuleMetrics struct {
	RuleID   string `json:"rule_id"`
	RuleName string `json:"rule_name"`
	Count    int64  `json:"count"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// LogEntry represents a log entry from VictoriaLogs
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
	Host      string                 `json:"host"`
	Status    int                    `json:"status"`
	RuleID    string                 `json:"rule_id,omitempty"`
	ClientIP  string                 `json:"client_ip"`
	Path      string                 `json:"path"`
	Method    string                 `json:"method"`
}

// LogQuery represents a log query
type LogQuery struct {
	Query     string    `json:"query"`
	TimeRange TimeRange `json:"time_range"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// LogSearchResult represents the result of a log search
type LogSearchResult struct {
	Entries    []LogEntry `json:"entries"`
	Total      int        `json:"total"`
	TimeRange  TimeRange  `json:"time_range"`
}

// AlertRule represents a vmalert rule
type AlertRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Expression  string    `json:"expression"`
	For         string    `json:"for"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Alert represents an active alert
type Alert struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	State        string    `json:"state"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	GeneratorURL string    `json:"generator_url"`
}

// AuditLog represents a configuration change audit log
type AuditLog struct {
	ID          string                 `json:"id"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	OldValue    interface{}            `json:"old_value"`
	NewValue    interface{}            `json:"new_value"`
	Diff        string                 `json:"diff"`
	User        string                 `json:"user"`
	Timestamp   time.Time              `json:"timestamp"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent"`
}

// WAFMode represents the WAF operating mode
type WAFMode string

const (
	WAFModeOn            WAFMode = "On"
	WAFModeDetectionOnly WAFMode = "DetectionOnly"
	WAFModeOff           WAFMode = "Off"
)

// PolicyUpdateRequest represents a policy update request
type PolicyUpdateRequest struct {
    Host        string        `json:"host" binding:"required"`
    Mode        string        `json:"mode" binding:"required,oneof=On DetectionOnly Off"`
    Namespace   string        `json:"namespace"`
    EnableCRS   *bool         `json:"enable_crs,omitempty"`
    Exceptions  *WAFExceptions `json:"exceptions,omitempty"`
    CustomRules []CustomRule  `json:"custom_rules,omitempty"`
}

// ExceptionUpdateRequest represents an exception update request
type ExceptionUpdateRequest struct {
    Host       string        `json:"host" binding:"required"`
    Namespace  string        `json:"namespace"`
    Exceptions WAFExceptions `json:"exceptions" binding:"required"`
    TestMode   bool          `json:"test_mode"`
}

// RuleUpdateRequest represents a rule update request
type RuleUpdateRequest struct {
    Host       string       `json:"host" binding:"required"`
    Namespace  string       `json:"namespace"`
    CustomRules []CustomRule `json:"custom_rules" binding:"required"`
}

// ApplyRequest represents a configuration apply request
type ApplyRequest struct {
    Host     string `json:"host" binding:"required"`
    Namespace string `json:"namespace"`
    Strategy string `json:"strategy" binding:"required,oneof=annotation configmap"`
}