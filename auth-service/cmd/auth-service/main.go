// Package main точка входа auth-service.
//
//	@title						Auth Service API
//	@version					1.0
//	@description				Микросервис авторизации и управления пользователями (мультитенантность Keycloak).
//	@host						localhost:8080
//	@BasePath					/
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/industrial-sed/auth-service/internal/config"
	"github.com/industrial-sed/auth-service/internal/handlers"
	"github.com/industrial-sed/auth-service/internal/jobs"
	"github.com/industrial-sed/auth-service/internal/jwtverify"
	kcclient "github.com/industrial-sed/auth-service/internal/keycloak"
	"github.com/industrial-sed/auth-service/internal/logger"
	appmigrate "github.com/industrial-sed/auth-service/internal/migrate"
	"github.com/industrial-sed/auth-service/internal/notifier"
	"github.com/industrial-sed/auth-service/internal/repositories"
	"github.com/industrial-sed/auth-service/internal/server"
	"github.com/industrial-sed/auth-service/internal/tracing"
	"github.com/industrial-sed/auth-service/internal/usecases"

	_ "github.com/industrial-sed/auth-service/docs"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		slog.Error("config", "error", err.Error())
		os.Exit(1)
	}
	log := logger.New(cfg.LogLevel)

	ctx := context.Background()
	shutdownTrace, err := tracing.Init(ctx, cfg.OTelEnabled, cfg.OTelServiceName, cfg.OTelExporterOTLPEndpoint)
	if err != nil {
		log.Error("otel_init", "error", err.Error())
		os.Exit(1)
	}
	defer func() { _ = shutdownTrace(context.Background()) }()

	pool, err := repositories.NewPool(ctx, cfg.DBDSN)
	if err != nil {
		log.Error("db", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.RunMigrationsOnStart {
		migDir := filepath.Join(".", "migrations")
		if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
			migDir = p
		}
		if err := appmigrate.Up(cfg.DBDSN, migDir); err != nil {
			log.Error("migrate", "error", err.Error())
			os.Exit(1)
		}
		log.Info("migrations_applied")
	}

	kc := kcclient.NewClient(
		cfg.KeycloakURL,
		cfg.KeycloakPublicURL,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.KeycloakClientSecret,
		cfg.KeycloakAdminRealm,
		cfg.KeycloakAdminUser,
		cfg.KeycloakAdminPassword,
	)

	admJWT, err := kc.LoginAdmin(ctx)
	if err != nil {
		log.Error("keycloak_admin_login", "error", err.Error())
		os.Exit(1)
	}
	admTok := admJWT.AccessToken
	if err := kc.EnsureRealmAndRoles(ctx, admTok); err != nil {
		log.Error("keycloak_bootstrap_realm", "error", err.Error())
		os.Exit(1)
	}
	cb := fmt.Sprintf("%s/api/v1/auth/callback", strings.TrimRight(strings.TrimSpace(cfg.APIPublicURL), "/"))
	postLogout := kcclient.BuildPostLogoutRedirectURIs(cfg.FrontendURL)
	if _, err := kc.EnsureOAuthClient(ctx, admTok, cb, postLogout); err != nil {
		log.Error("keycloak_bootstrap_client", "error", err.Error())
		os.Exit(1)
	}
	if err := kc.EnsureUserAttributeMapper(ctx, admTok); err != nil {
		log.Warn("keycloak_mapper", "error", err.Error())
	}
	if cfg.BootstrapSuperAdmin {
		if err := kcclient.BootstrapSuperAdmin(ctx, kc, cfg.BootstrapSuperAdminUsername, cfg.BootstrapSuperAdminPassword); err != nil {
			log.Error("bootstrap_superadmin", "error", err.Error())
			os.Exit(1)
		}
		log.Info("bootstrap_superadmin_ok", "user", cfg.BootstrapSuperAdminUsername)
	}

	n, err := notifier.New(cfg.NotifierType, log, cfg.KafkaBrokers, cfg.KafkaTopicUserEvents)
	if err != nil {
		log.Error("notifier", "error", err.Error())
		os.Exit(1)
	}

	tenantRepo := repositories.NewTenantRepo(pool)
	userRepo := repositories.NewUserCacheRepo(pool)
	tuc := usecases.NewTenantUC(kc, tenantRepo)
	uuc := usecases.NewUserUC(kc, tenantRepo, userRepo, n)

	// JWKS запрашиваем с URL, доступного из контейнера сервиса (внутренний Docker hostname).
	parser, err := jwtverify.NewParser(ctx, cfg.KeycloakURL, cfg.KeycloakRealm, cfg.KeycloakClientID)
	if err != nil {
		log.Error("jwt_parser", "error", err.Error())
		os.Exit(1)
	}

	authUC := usecases.NewAuthUC(
		kc,
		cfg.KeycloakPublicURL,
		cfg.KeycloakRealm,
		cfg.KeycloakClientID,
		cfg.APIPublicURL,
		cfg.FrontendURL,
		cfg.StateCookieSecret,
		userRepo,
	)

	tenantH := handlers.NewTenantHandler(tuc, uuc)
	userH := handlers.NewUserHandler(uuc)
	authH := handlers.NewAuthHandler(authUC, parser, cfg.CookieSecure())
	testH := handlers.NewTestHandler(kc, tenantRepo, userRepo)

	r := server.NewRouter(server.Deps{
		Config:   cfg,
		Log:      log,
		DB:       pool,
		Parser:   parser,
		TenantUC: tenantH,
		UserUC:   userH,
		Auth:     authH,
		Test:     testH,
		OTel:     cfg.OTelEnabled,
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: r}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go jobs.RunPasswordRotation(ctx, log, jobs.PasswordRotationConfig{
		Enabled: cfg.PasswordRotationEnabled,
		After:   cfg.PasswordRotationAfter,
		Every:   cfg.PasswordRotationEvery,
		Batch:   cfg.PasswordRotationBatch,
	}, kc, kc, userRepo)

	go func() {
		log.Info("listen", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout())
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Error("shutdown", "error", err.Error())
	}
}
