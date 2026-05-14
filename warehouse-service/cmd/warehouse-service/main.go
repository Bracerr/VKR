// Package main точка входа warehouse-service.
//
//	@title			Warehouse Service API
//	@version		1.0
//	@description	Склад: справочники, операции, резервы, отчёты
//	@host			localhost:8090
//	@BasePath		/api/v1
//	@securityDefinitions.apikey BearerAuth
//	@in				header
//	@name			Authorization
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/industrial-sed/warehouse-service/internal/config"
	"github.com/industrial-sed/warehouse-service/internal/clients"
	"github.com/industrial-sed/warehouse-service/internal/jobs"
	"github.com/industrial-sed/warehouse-service/internal/jwtverify"
	"github.com/industrial-sed/warehouse-service/internal/logger"
	appmigrate "github.com/industrial-sed/warehouse-service/internal/migrate"
	"github.com/industrial-sed/warehouse-service/internal/repositories"
	"github.com/industrial-sed/warehouse-service/internal/server"
	"github.com/industrial-sed/warehouse-service/internal/usecases"

	_ "github.com/industrial-sed/warehouse-service/docs"
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

	parser, err := jwtverify.NewParser(ctx, cfg.KeycloakURL, cfg.KeycloakRealm, cfg.KeycloakClientID)
	if err != nil {
		log.Error("jwks", "error", err.Error())
		os.Exit(1)
	}

	store := repositories.NewStore(pool)
	trace := clients.NewTraceability(cfg)
	uc := &usecases.UC{Store: store, DefaultCurrency: cfg.DefaultCurrency, Trace: trace}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	go jobs.RunReservationExpirer(bgCtx, log, uc, 2*time.Minute)
	go jobs.RunExpiryAlerts(bgCtx, log, store, 6*time.Hour, 30)

	r := server.NewRouter(server.Deps{
		Log:    log,
		Parser: parser,
		UC:     uc,
		Cfg:    cfg,
		DB:     pool,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("listen", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http", "error", err.Error())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	bgCancel()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ShutdownTimeout())*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown", "error", err.Error())
	}
}
