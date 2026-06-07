// Package main точка входа production-service.
//
//	@title			Production Service API
//	@version		1.0
//	@description	Производственный учёт (MES), BOM, техкарты, заказы, интеграция со складом и СЭД
//	@host			localhost:8092
//	@BasePath		/
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

	pcfg "github.com/industrial-sed/platform/config"
	platformkafka "github.com/industrial-sed/platform/kafka"
	"github.com/industrial-sed/platform/outbox"
	"github.com/industrial-sed/platform/publish"
	"github.com/industrial-sed/platform/runtime"
	"github.com/industrial-sed/platform/transport"

	"github.com/industrial-sed/production-service/internal/clients"
	"github.com/industrial-sed/production-service/internal/config"
	"github.com/industrial-sed/production-service/internal/handlers"
	"github.com/industrial-sed/production-service/internal/jwtverify"
	"github.com/industrial-sed/production-service/internal/logger"
	appmigrate "github.com/industrial-sed/production-service/internal/migrate"
	"github.com/industrial-sed/production-service/internal/repositories"
	"github.com/industrial-sed/production-service/internal/server"
	"github.com/industrial-sed/production-service/internal/usecases"

	_ "github.com/industrial-sed/production-service/docs"
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
	wh := clients.NewWarehouse(cfg)
	sed := clients.NewSED(cfg)
	trace := clients.NewTraceability(cfg)
	eventMode := transport.Mode(cfg.EventTransport)
	if v := pcfg.EventTransport(); v != "" {
		eventMode = transport.Mode(v)
	}
	brokers := cfg.KafkaBrokers
	if len(brokers) == 0 {
		brokers = pcfg.ParseBrokers("KAFKA_BROKERS", nil)
	}
	tracePub := &publish.TracePublisher{Mode: eventMode, HTTP: trace, Outbox: outbox.NewStore(pool)}
	appUC := &usecases.App{Store: store, WH: wh, SED: sed, Trace: trace, TracePub: tracePub, Cfg: cfg}
	httpHandlers := &handlers.HTTP{App: appUC}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	runtime.StartOutboxRelay(bgCtx, pool, brokers, string(eventMode), log)
	if transport.UseKafka(eventMode) && len(brokers) > 0 {
		platformkafka.StartSedSignedConsumer(bgCtx, brokers, "production-service", appUC.HandleSedDocumentSigned, log)
	}

	r := server.NewRouter(server.Deps{
		Log:    log,
		Parser: parser,
		App:    httpHandlers,
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
