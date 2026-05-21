package publish

import (
	"context"
	"log/slog"
	"time"

	"github.com/industrial-sed/platform/events"
	"github.com/industrial-sed/platform/outbox"
	"github.com/industrial-sed/platform/transport"
)

// TraceHTTPClient синхронная доставка в traceability (legacy HTTP).
type TraceHTTPClient interface {
	PostIngest(ctx context.Context, tenant string, legacy events.TraceIngestLegacy) error
}

// TracePublisher kafka/outbox + optional HTTP.
type TracePublisher struct {
	Mode   string
	Outbox *outbox.Store
	HTTP   TraceHTTPClient
}

func (p *TracePublisher) PublishDocumentPosted(ctx context.Context, tenant string, payload any, idemKey string) {
	p.publish(ctx, tenant, events.TypeTraceDocumentPosted, idemKey, payload)
}

func (p *TracePublisher) PublishLinkEntity(ctx context.Context, tenant string, payload events.LinkEntityWarehouseDocPayload, idemKey string) {
	p.publish(ctx, tenant, events.TypeTraceLinkEntityWhDoc, idemKey, payload)
}

func (p *TracePublisher) publish(ctx context.Context, tenant, eventType, idemKey string, payload any) {
	if p == nil {
		return
	}
	env, err := events.NewEnvelope(eventType, tenant, idemKey, payload)
	if err != nil {
		slog.Error("trace_outbox_build", "error", err.Error())
		return
	}
	if transport.UseKafka(p.Mode) && p.Outbox != nil {
		if err := p.Outbox.Enqueue(ctx, events.TopicTraceIngest, idemKey, env); err != nil {
			slog.Error("trace_outbox_enqueue", "error", err.Error())
		}
	}
	if transport.UseHTTP(p.Mode) && p.HTTP != nil {
		legacy, err := env.ToTraceIngest()
		if err != nil {
			slog.Error("trace_legacy_map", "error", err.Error())
			return
		}
		bg, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		if err := p.HTTP.PostIngest(bg, tenant, legacy); err != nil {
			slog.Error("trace_http_publish", "error", err.Error())
		}
	}
}
