package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/industrial-sed/platform/events"
)

// PostIngest HTTP ingest для platform/publish.
func (t *Traceability) PostIngest(ctx context.Context, tenant string, legacy events.TraceIngestLegacy) error {
	if !t.enabled() {
		return nil
	}
	legacy.TenantCode = tenant
	b, err := json.Marshal(legacy)
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
