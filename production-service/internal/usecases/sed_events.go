package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// HandleSedDocumentSigned callback после подписания документа в СЭД (X-Service-Secret).
func (a *App) HandleSedDocumentSigned(ctx context.Context, tenant string, documentID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	bom, err := a.Store.FindBOMBySedDocument(ctx, tx, tenant, documentID)
	if err != nil {
		return err
	}
	if bom != nil {
		if bom.Status == "APPROVED" {
			return nil
		}
		if bom.Status != "SUBMITTED" {
			return fmt.Errorf("%w: состояние BOM", ErrWrongState)
		}
		if err := a.Store.SetBOMStatusSED(ctx, tx, tenant, bom.ID, "APPROVED", bom.SedDocumentID); err != nil {
			return err
		}
		if err := a.Store.InsertHistory(ctx, tx, tenant, "bom", bom.ID, "sed_callback", "APPROVED_SED", histPayload(map[string]any{"document_id": documentID.String()})); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	rt, err := a.Store.FindRoutingBySedDocument(ctx, tx, tenant, documentID)
	if err != nil {
		return err
	}
	if rt != nil {
		if rt.Status == "APPROVED" {
			return nil
		}
		if rt.Status != "SUBMITTED" {
			return fmt.Errorf("%w: состояние маршрута", ErrWrongState)
		}
		if err := a.Store.SetRoutingStatusSED(ctx, tx, tenant, rt.ID, "APPROVED", rt.SedDocumentID); err != nil {
			return err
		}
		if err := a.Store.InsertHistory(ctx, tx, tenant, "routing", rt.ID, "sed_callback", "APPROVED_SED", histPayload(map[string]any{"document_id": documentID.String()})); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	return ErrNotFound
}
