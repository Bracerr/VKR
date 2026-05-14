package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

)

const (
	EventDocumentPosted = "DocumentPosted"
	EventLinkEntityDoc  = "LinkEntityWarehouseDoc"
)

type IngestEvent struct {
	EventType      string          `json:"event_type"`
	TenantCode     string          `json:"tenant_code"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Payload        json.RawMessage `json:"payload"`
}

type DocumentPostedPayload struct {
	DocumentID string    `json:"document_id"`
	DocType    string    `json:"doc_type"`
	Number     string    `json:"number,omitempty"`
	PostedAt   time.Time `json:"posted_at"`
	Lines      []struct {
		ProductID string  `json:"product_id"`
		BatchID   *string `json:"batch_id,omitempty"`
		BatchSeries *string `json:"batch_series,omitempty"`
		SerialID  *string `json:"serial_id,omitempty"`
		SerialNo  *string `json:"serial_no,omitempty"`
		Qty       string  `json:"qty"`
	} `json:"lines"`
}

type LinkEntityDocPayload struct {
	EntityType        string `json:"entity_type"` // SO/PO/PROD_ORDER/...
	EntityID          string `json:"entity_id"`
	EntityNumber      string `json:"entity_number,omitempty"`
	WarehouseDocumentID string `json:"warehouse_document_id"`
}

func (a *App) Ingest(ctx context.Context, ev *IngestEvent) error {
	if ev == nil || ev.TenantCode == "" || ev.EventType == "" {
		return ErrValidation
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := a.Store.InsertEvent(ctx, tx, ev.TenantCode, ev.EventType, ev.IdempotencyKey, ev.Payload); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// idempotency_key conflict → идемпотентный успех
			return tx.Commit(ctx)
		}
		return err
	}

	switch ev.EventType {
	case EventDocumentPosted:
		if err := a.applyDocumentPosted(ctx, tx, ev.TenantCode, ev.Payload); err != nil {
			return err
		}
	case EventLinkEntityDoc:
		if err := a.applyLinkEntityDoc(ctx, tx, ev.TenantCode, ev.Payload); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: unknown event_type", ErrValidation)
	}

	return tx.Commit(ctx)
}

func (a *App) applyDocumentPosted(ctx context.Context, tx pgx.Tx, tenant string, raw json.RawMessage) error {
	var p DocumentPostedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.DocumentID == "" || p.DocType == "" {
		return ErrValidation
	}
	docLabel := p.DocType
	if p.Number != "" {
		docLabel = p.DocType + " " + p.Number
	}
	docNodeID, err := a.Store.UpsertNode(ctx, tx, tenant, "WAREHOUSE_DOC", p.DocumentID, &docLabel, nil)
	if err != nil {
		return err
	}

	for _, ln := range p.Lines {
		// batch
		if ln.BatchID != nil && *ln.BatchID != "" {
			bLabel := "Batch"
			if ln.BatchSeries != nil && *ln.BatchSeries != "" {
				bLabel = "Batch " + *ln.BatchSeries
			}
			bMeta, _ := json.Marshal(map[string]any{
				"product_id": ln.ProductID,
				"series":     ln.BatchSeries,
			})
			bID, err := a.Store.UpsertNode(ctx, tx, tenant, "BATCH", *ln.BatchID, &bLabel, bMeta)
			if err != nil {
				return err
			}
			if err := a.Store.UpsertEdge(ctx, tx, tenant, "DOC_HAS_BATCH", docNodeID, bID, nil); err != nil {
				return err
			}
		}
		// serial
		if ln.SerialID != nil && *ln.SerialID != "" {
			sLabel := "Serial"
			if ln.SerialNo != nil && *ln.SerialNo != "" {
				sLabel = "SN " + *ln.SerialNo
			}
			sMeta, _ := json.Marshal(map[string]any{
				"product_id": ln.ProductID,
				"serial_no":  ln.SerialNo,
			})
			sID, err := a.Store.UpsertNode(ctx, tx, tenant, "SERIAL", *ln.SerialID, &sLabel, sMeta)
			if err != nil {
				return err
			}
			if err := a.Store.UpsertEdge(ctx, tx, tenant, "DOC_HAS_SERIAL", docNodeID, sID, nil); err != nil {
				return err
			}
			// serial -> batch (если есть)
			if ln.BatchID != nil && *ln.BatchID != "" {
				bID, _ := a.Store.GetNodeID(ctx, tx, tenant, "BATCH", *ln.BatchID)
				if bID != nil {
					_ = a.Store.UpsertEdge(ctx, tx, tenant, "SERIAL_IN_BATCH", sID, *bID, nil)
				}
			}
		}
	}
	return nil
}

func (a *App) applyLinkEntityDoc(ctx context.Context, tx pgx.Tx, tenant string, raw json.RawMessage) error {
	var p LinkEntityDocPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	if p.EntityType == "" || p.EntityID == "" || p.WarehouseDocumentID == "" {
		return ErrValidation
	}
	lbl := p.EntityType
	if p.EntityNumber != "" {
		lbl = p.EntityType + " " + p.EntityNumber
	}
	entNodeID, err := a.Store.UpsertNode(ctx, tx, tenant, p.EntityType, p.EntityID, &lbl, nil)
	if err != nil {
		return err
	}
	docNodeID, err := a.Store.UpsertNode(ctx, tx, tenant, "WAREHOUSE_DOC", p.WarehouseDocumentID, nil, nil)
	if err != nil {
		return err
	}
	return a.Store.UpsertEdge(ctx, tx, tenant, "ENTITY_POSTED_AS_DOC", entNodeID, docNodeID, nil)
}

