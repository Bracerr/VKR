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

	"github.com/industrial-sed/warehouse-service/internal/config"
)

// Traceability клиент к traceability-service (internal ingest).
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

// DocumentPostedEvent отправка события DocumentPosted.
func (t *Traceability) DocumentPostedEvent(ctx context.Context, tenant string, payload any, idemKey string) error {
	if !t.enabled() {
		return nil
	}
	body := map[string]any{
		"event_type":       "DocumentPosted",
		"tenant_code":      tenant,
		"idempotency_key":  idemKey,
		"payload":          payload,
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
		return fmt.Errorf("trace ingest: %d %s", resp.StatusCode, string(raw))
	}
	return nil
}

