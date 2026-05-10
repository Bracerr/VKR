// Package config загружает YAML и переменные окружения.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config настройки warehouse-service.
type Config struct {
	HTTPAddr string `mapstructure:"http_addr"`
	LogLevel string `mapstructure:"log_level"`

	DBDSN string `mapstructure:"db_dsn"`

	KeycloakURL       string `mapstructure:"keycloak_url"`
	KeycloakRealm     string `mapstructure:"keycloak_realm"`
	KeycloakClientID  string `mapstructure:"keycloak_client_id"`

	DefaultCurrency string `mapstructure:"default_currency"`

	RunMigrationsOnStart bool `mapstructure:"run_migrations_on_start"`

	RateLimitPerMinute int `mapstructure:"rate_limit_per_minute"`
	ImportMaxRows      int `mapstructure:"import_max_rows"`

	// ServiceSecret общий секрет для вызовов от sed-service (заголовок X-Service-Secret). Пусто — режим отключён.
	ServiceSecret string `mapstructure:"service_secret"`
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
	v.SetDefault("http_addr", ":8090")
	v.SetDefault("log_level", "info")
	v.SetDefault("db_dsn", "postgres://wh:wh@localhost:5433/warehouse?sslmode=disable")
	v.SetDefault("keycloak_url", "http://localhost:8081")
	v.SetDefault("keycloak_realm", "industrial-sed")
	v.SetDefault("keycloak_client_id", "auth-service")
	v.SetDefault("default_currency", "RUB")
	v.SetDefault("run_migrations_on_start", true)
	v.SetDefault("rate_limit_per_minute", 120)
	v.SetDefault("import_max_rows", 10000)
	v.SetDefault("service_secret", "")
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
	if c.DefaultCurrency == "" {
		c.DefaultCurrency = "RUB"
	}
	if c.RateLimitPerMinute <= 0 {
		c.RateLimitPerMinute = 120
	}
	if c.ImportMaxRows <= 0 {
		c.ImportMaxRows = 10000
	}
	return nil
}

// ShutdownTimeout graceful shutdown.
func ShutdownTimeout() int {
	return 15
}
