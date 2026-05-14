// Package config загружает YAML и переменные окружения.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config настройки procurement-service.
type Config struct {
	HTTPAddr string `mapstructure:"http_addr"`
	LogLevel string `mapstructure:"log_level"`

	DBDSN string `mapstructure:"db_dsn"`

	KeycloakURL      string `mapstructure:"keycloak_url"`
	KeycloakRealm    string `mapstructure:"keycloak_realm"`
	KeycloakClientID string `mapstructure:"keycloak_client_id"`

	RunMigrationsOnStart bool `mapstructure:"run_migrations_on_start"`
	RateLimitPerMinute   int  `mapstructure:"rate_limit_per_minute"`

	WarehouseBaseURL       string `mapstructure:"warehouse_base_url"`
	WarehouseServiceSecret string `mapstructure:"warehouse_service_secret"`

	SedBaseURL              string `mapstructure:"sed_base_url"`
	SedCallbackVerifySecret string `mapstructure:"sed_callback_verify_secret"`

	TraceabilityBaseURL string `mapstructure:"traceability_base_url"`
	TraceabilitySecret  string `mapstructure:"traceability_secret"`
}

// Load читает конфиг.
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	if configPath != "" {
		v.SetConfigFile(configPath)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}
	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("http_addr", ":8093")
	v.SetDefault("log_level", "info")
	v.SetDefault("db_dsn", "postgres://proc:proc@localhost:5436/procurement?sslmode=disable")
	v.SetDefault("keycloak_url", "http://localhost:8081")
	v.SetDefault("keycloak_realm", "industrial-sed")
	v.SetDefault("keycloak_client_id", "auth-service")
	v.SetDefault("run_migrations_on_start", true)
	v.SetDefault("rate_limit_per_minute", 120)
	v.SetDefault("warehouse_base_url", "http://localhost:8090")
	v.SetDefault("warehouse_service_secret", "")
	v.SetDefault("sed_base_url", "http://localhost:8091")
	v.SetDefault("sed_callback_verify_secret", "")
	v.SetDefault("traceability_base_url", "")
	v.SetDefault("traceability_secret", "")
}

// Validate проверяет обязательные поля.
func (c *Config) Validate() error {
	if c.DBDSN == "" {
		return errors.New("db_dsn is required")
	}
	if c.KeycloakURL == "" {
		return errors.New("keycloak_url is required")
	}
	if c.KeycloakRealm == "" {
		return errors.New("keycloak_realm is required")
	}
	if c.KeycloakClientID == "" {
		return errors.New("keycloak_client_id is required")
	}
	if c.RateLimitPerMinute <= 0 {
		c.RateLimitPerMinute = 120
	}
	if c.WarehouseBaseURL == "" {
		return errors.New("warehouse_base_url is required")
	}
	if c.SedBaseURL == "" {
		return errors.New("sed_base_url is required")
	}
	return nil
}

// ShutdownTimeout graceful shutdown.
func ShutdownTimeout() int { return 15 }

