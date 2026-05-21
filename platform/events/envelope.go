package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const SpecVersion = "1"

const (
	TypeSedDocumentSigned       = "sed.document.signed"
	TypeTraceDocumentPosted     = "trace.document_posted"
	TypeTraceLinkEntityWhDoc    = "trace.link_entity_warehouse_doc"
	TypeAuthUserCreated         = "auth.user.created"
)

const (
	TopicSedDocumentSigned = "vkr.sed.document.signed"
	TopicTraceIngest         = "vkr.trace.ingest"
	TopicAuthUserCreated     = "vkr.auth.user.created"
)

// Envelope единый формат события для Kafka.
type Envelope struct {
	SpecVersion    string          `json:"spec_version"`
	EventID        string          `json:"event_id"`
	EventType      string          `json:"event_type"`
	TenantCode     string          `json:"tenant_code"`
	OccurredAt     time.Time       `json:"occurred_at"`
	IdempotencyKey string          `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload"`
}

// NewEnvelope создаёт envelope с новым event_id.
func NewEnvelope(eventType, tenant, idemKey string, payload any) (Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return Envelope{}, err
		}
		raw = b
	}
	return Envelope{
		SpecVersion:    SpecVersion,
		EventID:        uuid.New().String(),
		EventType:      eventType,
		TenantCode:     tenant,
		OccurredAt:     time.Now().UTC(),
		IdempotencyKey: idemKey,
		Payload:        raw,
	}, nil
}

func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func Unmarshal(data []byte) (Envelope, error) {
	var e Envelope
	err := json.Unmarshal(data, &e)
	return e, err
}

// Headers для Kafka.
func (e Envelope) Headers() map[string]string {
	return map[string]string{
		"event_type":       e.EventType,
		"tenant_code":      e.TenantCode,
		"idempotency_key":  e.IdempotencyKey,
		"spec_version":     e.SpecVersion,
	}
}

// SedDocumentSignedPayload payload для sed.document.signed.
type SedDocumentSignedPayload struct {
	DocumentID       string `json:"document_id"`
	DocumentTypeCode string `json:"document_type_code"`
}

// LinkEntityWarehouseDocPayload связь домена со складским документом.
type LinkEntityWarehouseDocPayload struct {
	EntityType          string `json:"entity_type"`
	EntityID            string `json:"entity_id"`
	EntityNumber        string `json:"entity_number,omitempty"`
	WarehouseDocumentID string `json:"warehouse_document_id"`
}

// AuthUserCreatedPayload создание пользователя.
type AuthUserCreatedPayload struct {
	TenantCode        string `json:"tenant_code"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporary_password,omitempty"`
	KeycloakUserID    string `json:"keycloak_user_id"`
}

// TraceIngestLegacy — формат для usecases.Ingest (HTTP-совместимость).
type TraceIngestLegacy struct {
	EventType      string          `json:"event_type"`
	TenantCode     string          `json:"tenant_code"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Payload        json.RawMessage `json:"payload"`
}

// ToTraceIngest маппинг envelope → legacy ingest (DocumentPosted / LinkEntityWarehouseDoc).
func (e Envelope) ToTraceIngest() (TraceIngestLegacy, error) {
	legacyType := ""
	switch e.EventType {
	case TypeTraceDocumentPosted:
		legacyType = "DocumentPosted"
	case TypeTraceLinkEntityWhDoc:
		legacyType = "LinkEntityWarehouseDoc"
	default:
		return TraceIngestLegacy{}, ErrUnsupportedEventType
	}
	idem := e.IdempotencyKey
	return TraceIngestLegacy{
		EventType:      legacyType,
		TenantCode:     e.TenantCode,
		IdempotencyKey: &idem,
		Payload:        e.Payload,
	}, nil
}
