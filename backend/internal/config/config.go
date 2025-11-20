package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig   `mapstructure:"server"`
	Kubernetes K8sConfig      `mapstructure:"kubernetes"`
	Metrics    MetricsConfig  `mapstructure:"metrics"`
	Logs       LogsConfig     `mapstructure:"logs"`
	Security   SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Host         string   `mapstructure:"host"`
	Mode         string   `mapstructure:"mode"`
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type K8sConfig struct {
    ConfigPath     string `mapstructure:"config_path"`
    Namespace      string `mapstructure:"namespace"`
    ServiceAccount string `mapstructure:"service_account"`
    IngressControllerNamespace string `mapstructure:"ingress_controller_namespace"`
    IngressControllerConfigMapName string `mapstructure:"ingress_controller_configmap_name"`
    IngressControllerDeploymentName string `mapstructure:"ingress_controller_deployment_name"`
    WAFPoliciesConfigMapName string `mapstructure:"waf_policies_configmap_name"`
    DefaultIngressNamespace string `mapstructure:"default_ingress_namespace"`
    DefaultBackendServices []string `mapstructure:"default_backend_services"`
    DefaultApplyStrategy string `mapstructure:"default_apply_strategy"`
}

type MetricsConfig struct {
	VictoriaMetricsURL string `mapstructure:"victoria_metrics_url"`
	VmalertURL         string `mapstructure:"vmalert_url"`
}

type LogsConfig struct {
	VictoriaLogsURL string `mapstructure:"victoria_logs_url"`
}

type SecurityConfig struct {
	EnableAuth bool   `mapstructure:"enable_auth"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
}

var GlobalConfig *Config

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/waf-admin/")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "release")
    viper.SetDefault("server.allow_origins", []string{"*"})
    viper.SetDefault("kubernetes.namespace", "monitoring")
    viper.SetDefault("kubernetes.ingress_controller_namespace", "ingress-nginx")
    viper.SetDefault("kubernetes.ingress_controller_configmap_name", "ingress-nginx-controller")
    viper.SetDefault("kubernetes.ingress_controller_deployment_name", "ingress-nginx-controller")
    viper.SetDefault("kubernetes.waf_policies_configmap_name", "waf-policies")
    viper.SetDefault("kubernetes.default_ingress_namespace", "default")
    viper.SetDefault("kubernetes.default_backend_services", []string{"echo-server", "ingress-nginx-defaultbackend"})
    viper.SetDefault("kubernetes.default_apply_strategy", "annotation")
	viper.SetDefault("metrics.victoria_metrics_url", "http://victoria-metrics:8428")
	viper.SetDefault("metrics.vmalert_url", "http://vmalert:8880")
	viper.SetDefault("logs.victoria_logs_url", "http://victoria-logs:9428")
	viper.SetDefault("security.enable_auth", true)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("WAF")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		log.Println("Config file not found, using defaults and environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	GlobalConfig = &config
	return &config, nil
}

func GetConfig() *Config {
	if GlobalConfig == nil {
		config, err := LoadConfig()
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}
		GlobalConfig = config
	}
	return GlobalConfig
}