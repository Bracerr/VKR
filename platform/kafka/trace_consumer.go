package kafka

import (
	"context"
	"log/slog"

	pcfg "github.com/industrial-sed/platform/config"
	"github.com/industrial-sed/platform/events"
)

// TraceIngestHandler вызывает legacy ingest.
type TraceIngestHandler func(ctx context.Context, legacy events.TraceIngestLegacy) error

func TraceIngestConsumerHandler(h TraceIngestHandler) Handler {
	return func(ctx context.Context, env events.Envelope) error {
		legacy, err := env.ToTraceIngest()
		if err != nil {
			return err
		}
		return h(ctx, legacy)
	}
}

// StartTraceIngestConsumer consumer traceability-service.
func StartTraceIngestConsumer(ctx context.Context, brokers []string, groupID string, h TraceIngestHandler, log *slog.Logger) *Consumer {
	c := NewConsumer(ConsumerConfig{
		Brokers:     brokers,
		Topic:       events.TopicTraceIngest,
		GroupID:     groupID,
		StartOffset: pcfg.KafkaConsumerStartOffset(),
	}, TraceIngestConsumerHandler(h), log)
	go c.Run(ctx)
	return c
}
