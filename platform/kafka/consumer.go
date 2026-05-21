package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	pcfg "github.com/industrial-sed/platform/config"
	"github.com/industrial-sed/platform/events"
)

func consumerStartOffset() int64 {
	return pcfg.KafkaConsumerStartOffset()
}

// Handler обработчик envelope.
type Handler func(ctx context.Context, env events.Envelope) error

// Consumer читает топик и вызывает handler.
type Consumer struct {
	Reader  *kafka.Reader
	Handler Handler
	Log     *slog.Logger
}

type ConsumerConfig struct {
	Brokers       []string
	Topic         string
	GroupID       string
	StartOffset   int64 // kafka.FirstOffset or kafka.LastOffset
}

func NewConsumer(cfg ConsumerConfig, h Handler, log *slog.Logger) *Consumer {
	start := kafka.LastOffset
	if cfg.StartOffset == kafka.FirstOffset {
		start = kafka.FirstOffset
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		StartOffset:    start,
		CommitInterval: time.Second,
	})
	return &Consumer{Reader: r, Handler: h, Log: log}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		m, err := c.Reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if c.Log != nil {
				c.Log.Warn("kafka_fetch", "error", err.Error())
			}
			time.Sleep(time.Second)
			continue
		}
		env, err := events.Unmarshal(m.Value)
		if err != nil {
			if c.Log != nil {
				c.Log.Error("kafka_decode", "error", err.Error())
			}
			_ = c.Reader.CommitMessages(ctx, m)
			continue
		}
		if err := c.Handler(ctx, env); err != nil {
			if c.Log != nil {
				c.Log.Error("kafka_handler", "event_type", env.EventType, "error", err.Error())
			}
			// не коммитим — повторная доставка
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err := c.Reader.CommitMessages(ctx, m); err != nil && c.Log != nil {
			c.Log.Warn("kafka_commit", "error", err.Error())
		}
	}
}

func (c *Consumer) Close() error {
	return c.Reader.Close()
}
