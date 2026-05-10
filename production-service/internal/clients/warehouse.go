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

	"github.com/industrial-sed/production-service/internal/config"
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

func (w *Warehouse) headers(tenant string) http.Header {
	h := make(http.Header)
	h.Set(HeaderServiceSecret, w.secret)
	h.Set(HeaderTenantID, tenant)
	h.Set("Content-Type", "application/json")
	return h
}

func (w *Warehouse) post(ctx context.Context, tenant, path string, body any) ([]byte, int, error) {
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
	req.Header = w.headers(tenant)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}

func (w *Warehouse) postEmpty(ctx context.Context, tenant, path string) ([]byte, int, error) {
	if w.secret == "" {
		return nil, 0, fmt.Errorf("warehouse service secret не задан")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.base+path, http.NoBody)
	if err != nil {
		return nil, 0, err
	}
	req.Header = w.headers(tenant)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return raw, resp.StatusCode, nil
}

// WarehousePayloadLine строка склада.
type WarehousePayloadLine struct {
	BinID     *string `json:"bin_id"`
	ProductID string  `json:"product_id"`
	Qty       string  `json:"qty"`
	BatchID   *string `json:"batch_id"`
	SerialNo  *string `json:"serial_no"`
	Reason    string  `json:"reason"`
	DocRef    string  `json:"doc_ref"`
}

// WarehousePayload корень payload для интеграции.
type WarehousePayload struct {
	WarehouseID    string                 `json:"warehouse_id"`
	DefaultBinID   *string                `json:"default_bin_id"`
	Lines          []WarehousePayloadLine `json:"lines"`
	ReservationIDs []string               `json:"reservation_ids"`
}

// CreateReservations создаёт резервы по строкам.
func (w *Warehouse) CreateReservations(ctx context.Context, tenant, userName string, p *WarehousePayload) ([]uuid.UUID, error) {
	wh, err := uuid.Parse(p.WarehouseID)
	if err != nil {
		return nil, err
	}
	var ids []uuid.UUID
	for i, ln := range p.Lines {
		binStr := ""
		if ln.BinID != nil && *ln.BinID != "" {
			binStr = *ln.BinID
		} else if p.DefaultBinID != nil {
			binStr = *p.DefaultBinID
		}
		if binStr == "" {
			return nil, fmt.Errorf("строка %d: bin_id", i)
		}
		binID, err := uuid.Parse(binStr)
		if err != nil {
			return nil, err
		}
		pid, err := uuid.Parse(ln.ProductID)
		if err != nil {
			return nil, err
		}
		body := map[string]any{
			"warehouse_id": wh.String(),
			"bin_id":       binID.String(),
			"product_id":   pid.String(),
			"qty":          ln.Qty,
			"reason":       ln.Reason,
			"doc_ref":      ln.DocRef,
		}
		if ln.BatchID != nil {
			body["batch_id"] = *ln.BatchID
		}
		if ln.SerialNo != nil {
			body["serial_no"] = *ln.SerialNo
		}
		raw, code, err := w.post(ctx, tenant, "/api/v1/reservations", body)
		if err != nil {
			return nil, err
		}
		if code < 200 || code >= 300 {
			return nil, fmt.Errorf("warehouse reservations: %d %s", code, string(raw))
		}
		var out struct {
			ID uuid.UUID `json:"id"`
		}
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
		ids = append(ids, out.ID)
	}
	return ids, nil
}

// ConsumeReservation потребляет один резерв.
func (w *Warehouse) ConsumeReservation(ctx context.Context, tenant string, id uuid.UUID) error {
	raw, code, err := w.postEmpty(ctx, tenant, fmt.Sprintf("/api/v1/reservations/%s/consume", id.String()))
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("warehouse consume %s: %d %s", id, code, string(raw))
	}
	return nil
}

// ReleaseReservation снимает резерв.
func (w *Warehouse) ReleaseReservation(ctx context.Context, tenant string, id uuid.UUID) error {
	raw, code, err := w.postEmpty(ctx, tenant, fmt.Sprintf("/api/v1/reservations/%s/release", id.String()))
	if err != nil {
		return err
	}
	if code < 200 || code >= 300 {
		return fmt.Errorf("warehouse release %s: %d %s", id, code, string(raw))
	}
	return nil
}

// Receipt приход готовой продукции.
func (w *Warehouse) Receipt(ctx context.Context, tenant, userName string, p *WarehousePayload) (uuid.UUID, error) {
	wh, err := uuid.Parse(p.WarehouseID)
	if err != nil {
		return uuid.Nil, err
	}
	binStr := ""
	if p.DefaultBinID != nil {
		binStr = *p.DefaultBinID
	}
	if binStr == "" && len(p.Lines) > 0 && p.Lines[0].BinID != nil {
		binStr = *p.Lines[0].BinID
	}
	binID, err := uuid.Parse(binStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("bin_id")
	}
	var lines []map[string]any
	for _, ln := range p.Lines {
		lmap := map[string]any{
			"product_id": ln.ProductID,
			"qty":        ln.Qty,
		}
		if ln.BatchID != nil {
			lmap["batch_id"] = *ln.BatchID
		}
		if ln.SerialNo != nil {
			lmap["serial_numbers"] = []string{*ln.SerialNo}
		}
		lines = append(lines, lmap)
	}
	body := map[string]any{
		"warehouse_id": wh.String(),
		"bin_id":       binID.String(),
		"lines":        lines,
	}
	raw, code, err := w.post(ctx, tenant, "/api/v1/operations/receipt", body)
	if err != nil {
		return uuid.Nil, err
	}
	if code < 200 || code >= 300 {
		return uuid.Nil, fmt.Errorf("warehouse receipt: %d %s", code, string(raw))
	}
	var out struct {
		DocumentID uuid.UUID `json:"document_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return uuid.Nil, err
	}
	return out.DocumentID, nil
}
