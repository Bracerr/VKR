package publish

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/industrial-sed/platform/events"
	"github.com/industrial-sed/platform/outbox"
	"github.com/industrial-sed/platform/transport"
)

// SedCallbackHTTP legacy HTTP callbacks после подписи.
type SedCallbackHTTP interface {
	NotifyDocumentSigned(ctx context.Context, tenant string, documentID uuid.UUID, typeCode string) error
}

// SedSignedPublisher outbox/kafka и опционально HTTP.
type SedSignedPublisher struct {
	Mode       string
	Outbox     *outbox.Store
	Production SedCallbackHTTP
	Procurement SedCallbackHTTP
	Sales      SedCallbackHTTP
}

func (p *SedSignedPublisher) PublishDocumentSigned(ctx context.Context, tenant string, documentID uuid.UUID, typeCode string) {
	if p == nil {
		return
	}
	idem := "sed-sign-" + documentID.String()
	payload := events.SedDocumentSignedPayload{
		DocumentID:       documentID.String(),
		DocumentTypeCode: typeCode,
	}
	if transport.UseKafka(p.Mode) && p.Outbox != nil {
		env, err := events.NewEnvelope(events.TypeSedDocumentSigned, tenant, idem, payload)
		if err != nil {
			slog.Error("sed_outbox_build", "error", err.Error())
		} else if err := p.Outbox.Enqueue(ctx, events.TopicSedDocumentSigned, idem, env); err != nil {
			slog.Error("sed_outbox_enqueue", "error", err.Error())
		}
	}
	if transport.UseHTTP(p.Mode) {
		if p.Production != nil {
			if err := p.Production.NotifyDocumentSigned(ctx, tenant, documentID, typeCode); err != nil {
				slog.Error("sed_callback_production", "error", err.Error())
			}
		}
		if p.Procurement != nil {
			if err := p.Procurement.NotifyDocumentSigned(ctx, tenant, documentID, typeCode); err != nil {
				slog.Error("sed_callback_procurement", "error", err.Error())
			}
		}
		if p.Sales != nil {
			if err := p.Sales.NotifyDocumentSigned(ctx, tenant, documentID, typeCode); err != nil {
				slog.Error("sed_callback_sales", "error", err.Error())
			}
		}
	}
}
