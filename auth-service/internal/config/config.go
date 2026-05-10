// Package config загружает настройки из config.yaml и переменных окружения.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config — конфигурация сервиса авторизации.
type Config struct {
	HTTPAddr string `mapstructure:"http_addr"`
	LogLevel string `mapstructure:"log_level"`

	DBDSN string `mapstructure:"db_dsn"`

	KeycloakURL string `mapstructure:"keycloak_url"`
	// KeycloakPublicURL issuer в JWT (часто совпадает с URL в браузере). Если пусто — берётся keycloak_url.
	KeycloakPublicURL     string `mapstructure:"keycloak_public_url"`
	KeycloakAdminRealm    string `mapstructure:"keycloak_admin_realm"`
	KeycloakAdminUser     string `mapstructure:"keycloak_admin_user"`
	KeycloakAdminPassword string `mapstructure:"keycloak_admin_password"`
	KeycloakRealm         string `mapstructure:"keycloak_realm"`
	KeycloakClientID      string `mapstructure:"keycloak_client_id"`
	KeycloakClientSecret  string `mapstructure:"keycloak_client_secret"`

	FrontendURL string `mapstructure:"frontend_url"`
	// APIPublicURL публичный URL сервиса (для OIDC redirect_uri callback).
	APIPublicURL string `mapstructure:"api_public_url"`

	ServiceSecret     string `mapstructure:"service_secret"`
	TestSecret        string `mapstructure:"test_secret"`
	StateCookieSecret string `mapstructure:"state_cookie_secret"`

	EnableTestEndpoints bool `mapstructure:"enable_test_endpoints"`

	BootstrapSuperAdmin         bool   `mapstructure:"bootstrap_superadmin"`
	BootstrapSuperAdminUsername string `mapstructure:"bootstrap_superadmin_username"`
	BootstrapSuperAdminPassword string `mapstructure:"bootstrap_superadmin_password"`

	NotifierType         string   `mapstructure:"notifier_type"`
	KafkaBrokers         []string `mapstructure:"kafka_brokers"`
	KafkaTopicUserEvents string   `mapstructure:"kafka_topic_user_events"`

	AuthLoginRateLimit int `mapstructure:"auth_login_rate_limit"`

	OTelEnabled              bool   `mapstructure:"otel_enabled"`
	OTelServiceName          string `mapstructure:"otel_service_name"`
	OTelExporterOTLPEndpoint string `mapstructure:"otel_exporter_otlp_endpoint"`

	RunMigrationsOnStart bool `mapstructure:"run_migrations_on_start"`

	// PasswordRotationEnabled включает принудительную смену пароля спустя N времени после создания пользователя.
	PasswordRotationEnabled bool          `mapstructure:"password_rotation_enabled"`
	PasswordRotationAfter   time.Duration `mapstructure:"password_rotation_after"`
	PasswordRotationEvery   time.Duration `mapstructure:"password_rotation_every"`
	PasswordRotationBatch   int           `mapstructure:"password_rotation_batch"`
}

// Load читает config.yaml (если есть) и переопределяет из ENV.
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
	v.SetDefault("http_addr", ":8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("db_dsn", "postgres://auth:auth@localhost:5432/auth?sslmode=disable")
	v.SetDefault("keycloak_url", "http://localhost:8081")
	// Важно: Viper может не прокинуть env-only значение в Unmarshal, если ключ не известен.
	// Поэтому регистрируем ключ явно (иначе KEYCLOAK_PUBLIC_URL будет игнорироваться).
	v.SetDefault("keycloak_public_url", "")
	v.SetDefault("keycloak_admin_realm", "master")
	v.SetDefault("keycloak_admin_user", "")
	v.SetDefault("keycloak_admin_password", "")
	v.SetDefault("keycloak_realm", "industrial-sed")
	v.SetDefault("keycloak_client_id", "auth-service")
	// Иначе KEYCLOAK_CLIENT_SECRET из ENV может не попасть в Unmarshal.
	v.SetDefault("keycloak_client_secret", "")
	v.SetDefault("service_secret", "")
	v.SetDefault("test_secret", "")
	v.SetDefault("state_cookie_secret", "")
	v.SetDefault("frontend_url", "http://localhost:3000")
	v.SetDefault("api_public_url", "http://localhost:8080")
	v.SetDefault("enable_test_endpoints", false)
	v.SetDefault("bootstrap_superadmin_password", "")
	v.SetDefault("notifier_type", "mock")
	v.SetDefault("kafka_topic_user_events", "user.events")
	v.SetDefault("auth_login_rate_limit", 30)
	v.SetDefault("otel_enabled", false)
	v.SetDefault("otel_service_name", "auth-service")
	v.SetDefault("otel_exporter_otlp_endpoint", "http://localhost:4318")
	v.SetDefault("run_migrations_on_start", true)

	// Принудительная смена пароля (опционально; по умолчанию выключено).
	v.SetDefault("password_rotation_enabled", false)
	v.SetDefault("password_rotation_after", 7*24*time.Hour)
	v.SetDefault("password_rotation_every", 10*time.Minute)
	v.SetDefault("password_rotation_batch", 200)
}

// Validate проверяет обязательные поля.
func (c *Config) Validate() error {
	if c.DBDSN == "" {
		return errors.New("db_dsn is required")
	}
	if c.KeycloakURL == "" {
		return errors.New("keycloak_url is required")
	}
	if c.KeycloakPublicURL == "" {
		c.KeycloakPublicURL = c.KeycloakURL
	}
	if err := validateKeycloakPublicURLForBrowser(c.KeycloakPublicURL); err != nil {
		return err
	}
	if c.KeycloakRealm == "" {
		return errors.New("keycloak_realm is required")
	}
	if c.KeycloakClientID == "" || c.KeycloakClientSecret == "" {
		return errors.New("keycloak_client_id and keycloak_client_secret are required")
	}
	if len(c.StateCookieSecret) < 32 {
		return errors.New("state_cookie_secret must be at least 32 bytes")
	}
	if c.ServiceSecret == "" {
		return errors.New("service_secret is required")
	}
	if c.EnableTestEndpoints && c.TestSecret == "" {
		return errors.New("test_secret is required when enable_test_endpoints is true")
	}
	nt := strings.ToLower(c.NotifierType)
	if nt != "mock" && nt != "kafka" {
		return fmt.Errorf("notifier_type must be mock or kafka, got %q", c.NotifierType)
	}
	if nt == "kafka" && len(c.KafkaBrokers) == 0 {
		return errors.New("kafka_brokers required when notifier_type=kafka")
	}
	if c.PasswordRotationEvery <= 0 {
		c.PasswordRotationEvery = 10 * time.Minute
	}
	if c.PasswordRotationAfter <= 0 {
		c.PasswordRotationAfter = 7 * 24 * time.Hour
	}
	if c.PasswordRotationBatch <= 0 {
		c.PasswordRotationBatch = 200
	}
	return nil
}

// CookieSecure — в проде true при HTTPS.
func (c *Config) CookieSecure() bool {
	return strings.HasPrefix(c.FrontendURL, "https://")
}

// validateKeycloakPublicURLForBrowser отсекает типичную ошибку Docker: KEYCLOAK_PUBLIC_URL совпал с
// внутренним hostname keycloak, который не резолвится в браузере → NXDOMAIN на /auth/login.
func validateKeycloakPublicURLForBrowser(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil
	}
	if u.Hostname() == "keycloak" {
		return fmt.Errorf("keycloak_public_url: hostname %q is not resolvable in the browser; set KEYCLOAK_PUBLIC_URL to the URL you open in the browser (e.g. http://localhost:8081), keep KEYCLOAK_URL for server-side Admin API (e.g. http://keycloak:8080 in Docker)", u.Hostname())
	}
	return nil
}

// ShutdownTimeout для graceful shutdown.
func ShutdownTimeout() time.Duration {
	return 15 * time.Second
}
