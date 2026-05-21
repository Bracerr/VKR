package runtime

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	pcfg "github.com/industrial-sed/platform/config"
	platformkafka "github.com/industrial-sed/platform/kafka"
	"github.com/industrial-sed/platform/outbox"
	"github.com/industrial-sed/platform/transport"
)

// StartOutboxRelay если transport kafka/dual и brokers заданы.
func StartOutboxRelay(ctx context.Context, pool *pgxpool.Pool, cfgBrokers []string, transportMode string, log *slog.Logger) {
	mode := transport.Mode(transportMode)
	if v := pcfg.EventTransport(); v != "" {
		mode = transport.Mode(v)
	}
	brokers := cfgBrokers
	if len(brokers) == 0 {
		brokers = pcfg.ParseBrokers("KAFKA_BROKERS", nil)
	}
	if !transport.UseKafka(mode) || len(brokers) == 0 {
		return
	}
	producer := platformkafka.NewProducer(brokers)
	relay := &outbox.RelayWorker{
		Store:    outbox.NewStore(pool),
		Producer: producer,
		Brokers:  brokers,
		Log:      log,
		Interval: 2 * time.Second,
		DLQTopic: "vkr.dlq",
	}
	go relay.Run(ctx)
	if log != nil {
		log.Info("outbox_relay_started", "transport", mode)
	}
}
