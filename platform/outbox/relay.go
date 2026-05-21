package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/industrial-sed/platform/kafka"
)

// RelayWorker читает outbox и публикует в Kafka.
type RelayWorker struct {
	Store    *Store
	Producer *kafka.Producer
	Brokers  []string
	Log      *slog.Logger
	Interval time.Duration
	DLQTopic string // опционально: vkr.dlq.<topic>
}

func (w *RelayWorker) Run(ctx context.Context) {
	if w.Interval <= 0 {
		w.Interval = 2 * time.Second
	}
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *RelayWorker) tick(ctx context.Context) {
	rows, err := w.Store.FetchPendingBatch(ctx, 50)
	if err != nil {
		if w.Log != nil {
			w.Log.Error("outbox_fetch", "error", err.Error())
		}
		return
	}
	for _, row := range rows {
		w.processOne(ctx, row)
	}
}

func (w *RelayWorker) processOne(ctx context.Context, row Row) {
	env, err := EnvelopeFromRow(row)
	if err != nil {
		w.fail(ctx, row.ID, row.Attempts+1, err.Error())
		return
	}
	if err := w.Producer.Publish(ctx, w.Brokers, row.Topic, env); err != nil {
		w.fail(ctx, row.ID, row.Attempts+1, err.Error())
		if w.DLQTopic != "" && row.Attempts+1 >= maxAttempts {
			_ = w.Producer.Publish(ctx, w.Brokers, w.DLQTopic+"."+row.Topic, env)
		}
		return
	}
	if err := w.Store.MarkDelivered(ctx, row.ID); err != nil && w.Log != nil {
		w.Log.Error("outbox_delivered", "id", row.ID.String(), "error", err.Error())
	}
}

func (w *RelayWorker) fail(ctx context.Context, id uuid.UUID, attempts int, msg string) {
	next := time.Now().Add(Backoff(attempts))
	_ = w.Store.MarkFailed(ctx, id, msg, attempts, next)
	if w.Log != nil {
		w.Log.Warn("outbox_publish_failed", "id", id.String(), "attempts", attempts, "error", msg)
	}
}
