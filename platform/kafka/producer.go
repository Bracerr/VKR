package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/industrial-sed/platform/events"
)

// Producer публикует envelope в Kafka.
type Producer struct {
	writers map[string]*kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{writers: make(map[string]*kafka.Writer)}
}

func (p *Producer) writer(topic string, brokers []string) *kafka.Writer {
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	p.writers[topic] = w
	return w
}

func (p *Producer) Publish(ctx context.Context, brokers []string, topic string, env events.Envelope) error {
	body, err := env.Marshal()
	if err != nil {
		return err
	}
	hdrs := make([]kafka.Header, 0, len(env.Headers()))
	for k, v := range env.Headers() {
		hdrs = append(hdrs, kafka.Header{Key: k, Value: []byte(v)})
	}
	return p.writer(topic, brokers).WriteMessages(ctx, kafka.Message{
		Key:     []byte(env.IdempotencyKey),
		Value:   body,
		Headers: hdrs,
	})
}

func (p *Producer) Close() error {
	var first error
	for _, w := range p.writers {
		if err := w.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
