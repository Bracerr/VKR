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

// ReceiptLine строка приходной операции.
type ReceiptLine struct {
	ProductID string `json:"product_id"`
	Qty       string `json:"qty"`
	BatchID   *string `json:"batch_id,omitempty"`
}

// ReceiptRequest запрос прихода.
type ReceiptRequest struct {
	WarehouseID string       `json:"warehouse_id"`
	BinID       string       `json:"bin_id"`
	Lines       []ReceiptLine `json:"lines"`
}

// Receipt выполняет приход (warehouse-service: POST /operations/receipt).
func (w *Warehouse) Receipt(ctx context.Context, tenant string, req *ReceiptRequest, idempotencyKey string) (uuid.UUID, error) {
	raw, code, err := w.post(ctx, tenant, "/api/v1/operations/receipt", req, idempotencyKey)
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

