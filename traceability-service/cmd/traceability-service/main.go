// Package main точка входа traceability-service.
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

	pcfg "github.com/industrial-sed/platform/config"
	"github.com/industrial-sed/platform/events"
	platformkafka "github.com/industrial-sed/platform/kafka"
	"github.com/industrial-sed/platform/transport"

	"github.com/industrial-sed/traceability-service/internal/config"
	"github.com/industrial-sed/traceability-service/internal/handlers"
	"github.com/industrial-sed/traceability-service/internal/jwtverify"
	"github.com/industrial-sed/traceability-service/internal/logger"
	appmigrate "github.com/industrial-sed/traceability-service/internal/migrate"
	"github.com/industrial-sed/traceability-service/internal/repositories"
	"github.com/industrial-sed/traceability-service/internal/server"
	"github.com/industrial-sed/traceability-service/internal/usecases"
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
	app := &usecases.App{Store: store, Cfg: cfg}
	h := &handlers.HTTP{App: app}
	r := server.NewRouter(server.Deps{Log: log, Parser: parser, H: h, Cfg: cfg, DB: pool})

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	eventMode := transport.Mode(cfg.EventTransport)
	if v := pcfg.EventTransport(); v != "" {
		eventMode = transport.Mode(v)
	}
	brokers := cfg.KafkaBrokers
	if len(brokers) == 0 {
		brokers = pcfg.ParseBrokers("KAFKA_BROKERS", nil)
	}
	if transport.UseKafka(eventMode) && len(brokers) > 0 {
		group := cfg.KafkaConsumerGroup
		if group == "" {
			group = "traceability-service"
		}
		platformkafka.StartTraceIngestConsumer(bgCtx, brokers, group, func(ctx context.Context, legacy events.TraceIngestLegacy) error {
			return app.Ingest(ctx, &usecases.IngestEvent{
				EventType:      legacy.EventType,
				TenantCode:     legacy.TenantCode,
				IdempotencyKey: legacy.IdempotencyKey,
				Payload:        legacy.Payload,
			})
		}, log)
		log.Info("kafka_trace_consumer_started")
	}

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

