package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/industrial-sed/sales-service/internal/config"
)

// Traceability client к traceability-service (internal events).
type Traceability struct {
	base   string
	secret string
	client *http.Client
}

func NewTraceability(cfg *config.Config) *Traceability {
	return &Traceability{
		base:   strings.TrimRight(cfg.TraceabilityBaseURL, "/"),
		secret: cfg.TraceabilitySecret,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *Traceability) enabled() bool {
	return t != nil && t.base != "" && t.secret != ""
}

func (t *Traceability) LinkEntityToWarehouseDoc(ctx context.Context, tenant, entityType, entityID, entityNumber, warehouseDocID, idemKey string) error {
	if !t.enabled() {
		return nil
	}
	payload := map[string]any{
		"entity_type":          entityType,
		"entity_id":            entityID,
		"entity_number":        entityNumber,
		"warehouse_document_id": warehouseDocID,
	}
	body := map[string]any{
		"event_type":      "LinkEntityWarehouseDoc",
		"tenant_code":     tenant,
		"idempotency_key": idemKey,
		"payload":         payload,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.base+"/api/v1/internal/events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Secret", t.secret)
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("trace link: %d %s", resp.StatusCode, string(raw))
	}
	return nil
}

