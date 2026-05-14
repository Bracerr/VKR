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

	"github.com/industrial-sed/sales-service/internal/config"
)

const (
	HeaderServiceSecret = "X-Service-Secret"
	HeaderTenantID      = "X-Tenant-Id"
)

// Warehouse HTTP-клиент к warehouse-service.
type Warehouse struct {
	base   string
	secret string
	client *http.Client
}

// NewWarehouse клиент.
func NewWarehouse(cfg *config.Config) *Warehouse {
	return &Warehouse{
		base:   strings.TrimRight(cfg.WarehouseBaseURL, "/"),
		secret: cfg.WarehouseServiceSecret,
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (w *Warehouse) headers(tenant, idemKey string) http.Header {
	h := make(http.Header)
	h.Set(HeaderServiceSecret, w.secret)
	h.Set(HeaderTenantID, tenant)
	h.Set("Content-Type", "application/json")
	if strings.TrimSpace(idemKey) != "" {
		h.Set("Idempotency-Key", idemKey)
	}
	return h
}

func (w *Warehouse) post(ctx context.Context, tenant, path string, body any, idemKey string) ([]byte, int, error) {
	if w.secret == "" {
		return nil, 0, fmt.Errorf("warehouse service secret не задан")
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.base+path, bytes.NewReader(b))
	if err != nil {
		return nil, 0, err
	}
	req.Header = w.headers(tenant, idemKey)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}

func (w *Warehouse) postEmpty(ctx context.Context, tenant, path string, idemKey string) ([]byte, int, error) {
	if w.secret == "" {
		return nil, 0, fmt.Errorf("warehouse service secret не задан")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.base+path, http.NoBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header = w.headers(tenant, idemKey)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}

// ReservationRequest создать резерв.
type ReservationRequest struct {
	WarehouseID string  `json:"warehouse_id"`
	BinID       string  `json:"bin_id"`
	ProductID   string  `json:"product_id"`
	BatchID     *string `json:"batch_id,omitempty"`
	SerialNo    *string `json:"serial_no,omitempty"`
	Qty         string  `json:"qty"`
	Reason      string  `json:"reason,omitempty"`
	DocRef      string  `json:"doc_ref,omitempty"`
}

// CreateReservation POST /api/v1/reservations.
func (w *Warehouse) CreateReservation(ctx context.Context, tenant string, req *ReservationRequest, idemKey string) (uuid.UUID, error) {
	raw, code, err := w.post(ctx, tenant, "/api/v1/reservations", req, idemKey)
	if err != nil {
		return uuid.Nil, err
	}
	if code < 200 || code >= 300 {
		return uuid.Nil, fmt.Errorf("warehouse reservation: %d %s", code, string(raw))
	}
	var out struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return uuid.Nil, err
	}
	return out.ID, nil
}

// ReleaseReservation POST /api/v1/reservations/:id/release.
func (w *Warehouse) ReleaseReservation(ctx context.Context, tenant string, id uuid.UUID) error {
	raw, code, err := w.postEmpty(ctx, tenant, "/api/v1/reservations/"+id.String()+"/release", "")
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("warehouse reservation release: %d %s", code, string(raw))
	}
	return nil
}

// IssueLine строка отгрузки (issue).
type IssueLine struct {
	ProductID string   `json:"product_id"`
	Qty       string   `json:"qty,omitempty"`
	BatchID   *string  `json:"batch_id,omitempty"`
	SerialNumbers []string `json:"serial_numbers,omitempty"`
}

type IssueRequest struct {
	WarehouseID string     `json:"warehouse_id"`
	BinID       string     `json:"bin_id"`
	Lines       []IssueLine `json:"lines"`
}

// Issue POST /api/v1/operations/issue.
func (w *Warehouse) Issue(ctx context.Context, tenant string, req *IssueRequest, idemKey string) (uuid.UUID, error) {
	raw, code, err := w.post(ctx, tenant, "/api/v1/operations/issue", req, idemKey)
	if err != nil {
		return uuid.Nil, err
	}
	if code < 200 || code >= 300 {
		return uuid.Nil, fmt.Errorf("warehouse issue: %d %s", code, string(raw))
	}
	var out struct {
		DocumentID uuid.UUID `json:"document_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return uuid.Nil, err
	}
	return out.DocumentID, nil
}

type IssueFromReservationsRequest struct {
	ReservationIDs []string `json:"reservation_ids"`
}

// IssueFromReservations POST /api/v1/operations/issue-from-reservations.
func (w *Warehouse) IssueFromReservations(ctx context.Context, tenant string, reservationIDs []uuid.UUID, idemKey string) (uuid.UUID, error) {
	var req IssueFromReservationsRequest
	for _, id := range reservationIDs {
		req.ReservationIDs = append(req.ReservationIDs, id.String())
	}
	raw, code, err := w.post(ctx, tenant, "/api/v1/operations/issue-from-reservations", &req, idemKey)
	if err != nil {
		return uuid.Nil, err
	}
	if code < 200 || code >= 300 {
		return uuid.Nil, fmt.Errorf("warehouse issue-from-reservations: %d %s", code, string(raw))
	}
	var out struct {
		DocumentID uuid.UUID `json:"document_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return uuid.Nil, err
	}
	return out.DocumentID, nil
}

