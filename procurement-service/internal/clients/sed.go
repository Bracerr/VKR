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

	"github.com/industrial-sed/procurement-service/internal/config"
)

// SED клиент к sed-service (от имени пользователя по Bearer).
type SED struct {
	base   string
	client *http.Client
}

// NewSED создаёт клиент.
func NewSED(cfg *config.Config) *SED {
	return &SED{
		base:   strings.TrimRight(cfg.SedBaseURL, "/"),
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

// SedDocument минимальный ответ создания.
type SedDocument struct {
	ID         uuid.UUID       `json:"id"`
	Status     string          `json:"status"`
	TenantCode string          `json:"tenant_code"`
	Payload    json.RawMessage `json:"payload"`
}

// CreateDocument POST /api/v1/documents.
func (s *SED) CreateDocument(ctx context.Context, bearer string, typeID uuid.UUID, title string, payload json.RawMessage) (*SedDocument, error) {
	if bearer == "" {
		return nil, fmt.Errorf("нет bearer для sed")
	}
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	body := struct {
		TypeID  string          `json:"type_id"`
		Title   string          `json:"title"`
		Payload json.RawMessage `json:"payload"`
	}{
		TypeID: typeID.String(), Title: title, Payload: payload,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.base+"/api/v1/documents", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("sed create document: %d %s", resp.StatusCode, string(raw))
	}
	var out SedDocument
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitDocument POST /documents/:id/submit.
func (s *SED) SubmitDocument(ctx context.Context, bearer string, docID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.base+"/api/v1/documents/"+docID.String()+"/submit", http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sed submit: %d %s", resp.StatusCode, string(raw))
	}
	return nil
}

