package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

func (u *UC) emitTraceDocumentPosted(ctx context.Context, tenant string, docID uuid.UUID) {
	if u.TracePub == nil {
		return
	}
	d, err := u.Store.GetDocument(ctx, nil, tenant, docID)
	if err != nil || d == nil {
		return
	}
	movs, err := u.Store.ListMovementsByDocument(ctx, nil, tenant, docID)
	if err != nil {
		return
	}

	type outLine struct {
		ProductID   string  `json:"product_id"`
		BatchID     *string `json:"batch_id,omitempty"`
		BatchSeries *string `json:"batch_series,omitempty"`
		SerialID    *string `json:"serial_id,omitempty"`
		SerialNo    *string `json:"serial_no,omitempty"`
		Qty         string  `json:"qty"`
	}
	payload := struct {
		DocumentID string    `json:"document_id"`
		DocType    string    `json:"doc_type"`
		Number     string    `json:"number,omitempty"`
		PostedAt   time.Time `json:"posted_at"`
		Lines      []outLine `json:"lines"`
	}{
		DocumentID: d.ID.String(),
		DocType:    d.DocType,
		Number:     d.Number,
		PostedAt:   d.CreatedAt.UTC(),
	}

	for _, mv := range movs {
		ol := outLine{ProductID: mv.ProductID.String(), Qty: mv.Qty}
		if mv.BatchID != models.NilBatchID {
			bid := mv.BatchID.String()
			ol.BatchID = &bid
			b, _ := u.Store.GetBatchByID(ctx, nil, tenant, mv.BatchID)
			if b != nil && b.Series != "" {
				ser := b.Series
				ol.BatchSeries = &ser
			}
		}
		if mv.SerialID != nil {
			sid := mv.SerialID.String()
			ol.SerialID = &sid
			sn, _ := u.Store.GetSerialByID(ctx, nil, tenant, *mv.SerialID)
			if sn != nil && sn.SerialNo != "" {
				s := sn.SerialNo
				ol.SerialNo = &s
			}
		}
		payload.Lines = append(payload.Lines, ol)
	}

	u.TracePub.PublishDocumentPosted(ctx, tenant, payload, "wh-doc-"+docID.String())
}
