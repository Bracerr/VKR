// Package notifier отправляет события о пользователях (mock или Kafka).
package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"

	"github.com/industrial-sed/auth-service/internal/usecases"
)

// New создаёт notifier по типу из конфига.
func New(typ string, log *slog.Logger, kafkaBrokers []string, topic string) (usecases.Notifier, error) {
	switch typ {
	case "kafka":
		if len(kafkaBrokers) == 0 || topic == "" {
			return nil, fmt.Errorf("kafka: brokers and topic required")
		}
		return NewKafkaNotifier(kafkaBrokers, topic, log), nil
	default:
		return NewMockNotifier(log), nil
	}
}

// MockNotifier пишет событие в лог (JSON).
type MockNotifier struct {
	log *slog.Logger
}

// NewMockNotifier конструктор.
func NewMockNotifier(log *slog.Logger) *MockNotifier {
	return &MockNotifier{log: log}
}

// NotifyUserCreated логирует payload.
func (m *MockNotifier) NotifyUserCreated(ctx context.Context, payload usecases.UserCreatedPayload) error {
	b, _ := json.Marshal(payload)
	m.log.Info("notify_user_created_mock", "payload", string(b))
	return nil
}

// KafkaNotifier публикует JSON в топик.
type KafkaNotifier struct {
	writer *kafka.Writer
	log    *slog.Logger
}

// NewKafkaNotifier конструктор.
func NewKafkaNotifier(brokers []string, topic string, log *slog.Logger) *KafkaNotifier {
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaNotifier{writer: w, log: log}
}

// NotifyUserCreated отправляет сообщение в Kafka.
func (k *KafkaNotifier) NotifyUserCreated(ctx context.Context, payload usecases.UserCreatedPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := k.writer.WriteMessages(ctx, kafka.Message{Value: b}); err != nil {
		k.log.Error("kafka_write_failed", "error", err.Error())
		return err
	}
	return nil
}
