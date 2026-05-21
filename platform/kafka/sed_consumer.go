package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/industrial-sed/platform/events"
)

// SedSignedHandler обработка sed.document.signed.
type SedSignedHandler func(ctx context.Context, tenant string, documentID uuid.UUID) error

func SedSignedConsumerHandler(h SedSignedHandler) Handler {
	return func(ctx context.Context, env events.Envelope) error {
		if env.EventType != events.TypeSedDocumentSigned {
			return nil
		}
		var p events.SedDocumentSignedPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		docID, err := uuid.Parse(p.DocumentID)
		if err != nil {
			return err
		}
		return h(ctx, env.TenantCode, docID)
	}
}

// StartSedSignedConsumer запуск consumer для доменного сервиса.
func StartSedSignedConsumer(ctx context.Context, brokers []string, groupID string, h SedSignedHandler, log *slog.Logger) *Consumer {
	c := NewConsumer(ConsumerConfig{
		Brokers:     brokers,
		Topic:       events.TopicSedDocumentSigned,
		GroupID:     groupID,
		StartOffset: consumerStartOffset(),
	}, SedSignedConsumerHandler(h), log)
	go c.Run(ctx)
	return c
}
