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

	"github.com/google/uuid"
)

// ProductionCallback уведомляет production-service о подписании документа.
type ProductionCallback struct {
	BaseURL string
	Secret  string
	Client  *http.Client
}

// NotifyDocumentSigned POST в production-service /api/v1/internal/sed-events.
func (p *ProductionCallback) NotifyDocumentSigned(ctx context.Context, tenant string, documentID uuid.UUID, typeCode string) error {
	if p == nil || strings.TrimSpace(p.BaseURL) == "" || strings.TrimSpace(p.Secret) == "" {
		return nil
	}
	cl := p.Client
	if cl == nil {
		cl = &http.Client{Timeout: 15 * time.Second}
	}
	base := strings.TrimRight(p.BaseURL, "/")
	body := map[string]any{
		"event":                "DOCUMENT_SIGNED",
		"tenant_code":          tenant,
		"document_id":          documentID.String(),
		"document_type_code":   typeCode,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v1/internal/sed-events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Secret", p.Secret)
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("production callback: %d %s", resp.StatusCode, string(raw))
	}
	return nil
}
