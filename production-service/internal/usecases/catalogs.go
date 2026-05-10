package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/production-service/internal/models"
)

func histPayload(m map[string]any) json.RawMessage {
	b, _ := json.Marshal(m)
	return b
}

// --- Workcenters ---

func (a *App) ListWorkcenters(ctx context.Context, tenant string) ([]models.Workcenter, error) {
	return a.Store.ListWorkcenters(ctx, nil, tenant)
}

func (a *App) CreateWorkcenter(ctx context.Context, tenant, actorSub, code, name string, active bool, cap *int) (*models.Workcenter, error) {
	w := &models.Workcenter{
		ID: uuid.New(), TenantCode: tenant, Code: code, Name: name, Active: active,
		CapacityMinutesPerShift: cap,
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.CreateWorkcenter(ctx, tx, w); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "workcenter", w.ID, actorSub, "CREATE", histPayload(map[string]any{"code": code})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (a *App) UpdateWorkcenter(ctx context.Context, tenant, actorSub string, id uuid.UUID, code, name string, active bool, cap *int) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	w, err := a.Store.GetWorkcenter(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if w == nil {
		return ErrNotFound
	}
	w.Code, w.Name, w.Active = code, name, active
	w.CapacityMinutesPerShift = cap
	if err := a.Store.UpdateWorkcenter(ctx, tx, w); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "workcenter", id, actorSub, "UPDATE", histPayload(nil)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (a *App) DeleteWorkcenter(ctx context.Context, tenant, actorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteWorkcenter(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "workcenter", id, actorSub, "DELETE", nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// --- Scrap ---

func (a *App) ListScrapReasons(ctx context.Context, tenant string) ([]models.ScrapReason, error) {
	return a.Store.ListScrapReasons(ctx, nil, tenant)
}

func (a *App) CreateScrapReason(ctx context.Context, tenant, actorSub, code, name string) (*models.ScrapReason, error) {
	r := &models.ScrapReason{ID: uuid.New(), TenantCode: tenant, Code: code, Name: name}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.CreateScrapReason(ctx, tx, r); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "scrap_reason", r.ID, actorSub, "CREATE", histPayload(map[string]any{"code": code})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

// --- BOM ---

func (a *App) ListBOMs(ctx context.Context, tenant string, productID *uuid.UUID, status *string) ([]models.BOM, error) {
	return a.Store.ListBOMs(ctx, nil, tenant, productID, status)
}

func (a *App) GetBOMWithLines(ctx context.Context, tenant string, id uuid.UUID) (*models.BOM, []models.BOMLine, error) {
	b, err := a.Store.GetBOM(ctx, nil, tenant, id)
	if err != nil {
		return nil, nil, err
	}
	if b == nil {
		return nil, nil, ErrNotFound
	}
	lines, err := a.Store.ListBOMLines(ctx, nil, id)
	if err != nil {
		return nil, nil, err
	}
	return b, lines, nil
}

func (a *App) CreateBOM(ctx context.Context, tenant, actorSub string, productID uuid.UUID) (*models.BOM, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	v, err := a.Store.NextBOMVersion(ctx, tx, tenant, productID)
	if err != nil {
		return nil, err
	}
	b := &models.BOM{
		ID: uuid.New(), TenantCode: tenant, ProductID: productID, Version: v,
		Status: "DRAFT",
	}
	if err := a.Store.CreateBOM(ctx, tx, b); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "bom", b.ID, actorSub, "CREATE", histPayload(map[string]any{"version": v})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return b, nil
}

func (a *App) UpdateBOMDates(ctx context.Context, tenant, actorSub string, id uuid.UUID, validFrom, validTo *time.Time) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	b, err := a.Store.GetBOM(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if b == nil {
		return ErrNotFound
	}
	if b.Status != "DRAFT" {
		return ErrWrongState
	}
	b.ValidFrom, b.ValidTo = validFrom, validTo
	if err := a.Store.UpdateBOM(ctx, tx, b); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "bom", id, actorSub, "UPDATE", histPayload(nil)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (a *App) AddBOMLine(ctx context.Context, tenant, actorSub string, bomID uuid.UUID, lineNo int, componentProductID uuid.UUID, qtyPer, scrapPct string, opNo int, altGroup *string) (*models.BOMLine, error) {
	qty, err := decimal.NewFromString(qtyPer)
	if err != nil || qty.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: qty_per", ErrValidation)
	}
	scr, err := decimal.NewFromString(scrapPct)
	if err != nil || scr.IsNegative() {
		return nil, fmt.Errorf("%w: scrap_pct", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	b, err := a.Store.GetBOM(ctx, tx, tenant, bomID)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrNotFound
	}
	if b.Status != "DRAFT" {
		return nil, ErrWrongState
	}
	if componentProductID == b.ProductID {
		return nil, fmt.Errorf("%w: компонент совпадает с изделием", ErrValidation)
	}
	if lineNo <= 0 {
		return nil, fmt.Errorf("%w: line_no", ErrValidation)
	}
	ln := &models.BOMLine{
		ID: uuid.New(), BomID: bomID, LineNo: lineNo, ComponentProductID: componentProductID,
		QtyPer: qty.StringFixed(8), ScrapPct: scr.StringFixed(4), OpNo: opNo, AltGroup: altGroup,
	}
	if err := a.Store.AddBOMLine(ctx, tx, ln); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "bom_line", ln.ID, actorSub, "ADD", histPayload(map[string]any{"bom_id": bomID.String()})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ln, nil
}

func (a *App) DeleteBOMLine(ctx context.Context, tenant, actorSub string, bomID, lineID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	b, err := a.Store.GetBOM(ctx, tx, tenant, bomID)
	if err != nil {
		return err
	}
	if b == nil {
		return ErrNotFound
	}
	if b.Status != "DRAFT" {
		return ErrWrongState
	}
	if err := a.Store.DeleteBOMLine(ctx, tx, bomID, lineID); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "bom", bomID, actorSub, "DELETE_LINE", histPayload(map[string]any{"line_id": lineID.String()})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SubmitBOM отправляет тип документа на согласование в СЭД (через Bearer пользователя).
func (a *App) SubmitBOM(ctx context.Context, tenant, actorSub, bearer string, bomID, sedDocumentTypeID uuid.UUID, title string) error {
	if bearer == "" {
		return fmt.Errorf("%w: Authorization для sed", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	b, err := a.Store.GetBOM(ctx, tx, tenant, bomID)
	if err != nil {
		return err
	}
	if b == nil {
		return ErrNotFound
	}
	if b.Status != "DRAFT" {
		return ErrWrongState
	}
	lines, err := a.Store.ListBOMLines(ctx, tx, bomID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("%w: пустая спецификация", ErrValidation)
	}
	payload, _ := json.Marshal(map[string]string{
		"kind":   "bom",
		"bom_id": bomID.String(),
	})
	doc, err := a.SED.CreateDocument(ctx, bearer, sedDocumentTypeID, title, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	if err := a.SED.SubmitDocument(ctx, bearer, doc.ID); err != nil {
		return fmt.Errorf("%w: sed submit: %v", ErrValidation, err)
	}
	sedID := doc.ID
	if err := a.Store.SetBOMStatusSED(ctx, tx, tenant, bomID, "SUBMITTED", &sedID); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "bom", bomID, actorSub, "SUBMIT_SED", histPayload(map[string]any{"sed_document_id": sedID.String()})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (a *App) ArchiveBOM(ctx context.Context, tenant, actorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	b, err := a.Store.GetBOM(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if b == nil {
		return ErrNotFound
	}
	if err := a.Store.SetBOMStatusSED(ctx, tx, tenant, id, "ARCHIVED", b.SedDocumentID); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "bom", id, actorSub, "ARCHIVE", nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// --- Routings ---

func (a *App) ListRoutings(ctx context.Context, tenant string, productID *uuid.UUID, status *string) ([]models.Routing, error) {
	return a.Store.ListRoutings(ctx, nil, tenant, productID, status)
}

func (a *App) GetRoutingWithOps(ctx context.Context, tenant string, id uuid.UUID) (*models.Routing, []models.RoutingOperation, error) {
	r, err := a.Store.GetRouting(ctx, nil, tenant, id)
	if err != nil {
		return nil, nil, err
	}
	if r == nil {
		return nil, nil, ErrNotFound
	}
	ops, err := a.Store.ListRoutingOperations(ctx, nil, id)
	if err != nil {
		return nil, nil, err
	}
	return r, ops, nil
}

func (a *App) CreateRouting(ctx context.Context, tenant, actorSub string, productID uuid.UUID) (*models.Routing, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	v, err := a.Store.NextRoutingVersion(ctx, tx, tenant, productID)
	if err != nil {
		return nil, err
	}
	r := &models.Routing{
		ID: uuid.New(), TenantCode: tenant, ProductID: productID, Version: v, Status: "DRAFT",
	}
	if err := a.Store.CreateRouting(ctx, tx, r); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "routing", r.ID, actorSub, "CREATE", histPayload(map[string]any{"version": v})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (a *App) AddRoutingOperation(ctx context.Context, tenant, actorSub string, routingID uuid.UUID, opNo int, workcenterID uuid.UUID, name string, tpu, stu *string, qc bool) (*models.RoutingOperation, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	r, err := a.Store.GetRouting(ctx, tx, tenant, routingID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrNotFound
	}
	if r.Status != "DRAFT" {
		return nil, ErrWrongState
	}
	wc, err := a.Store.GetWorkcenter(ctx, tx, tenant, workcenterID)
	if err != nil {
		return nil, err
	}
	if wc == nil {
		return nil, fmt.Errorf("%w: workcenter", ErrValidation)
	}
	op := &models.RoutingOperation{
		ID: uuid.New(), RoutingID: routingID, OpNo: opNo, WorkcenterID: workcenterID,
		Name: name, TimePerUnitMin: tpu, SetupTimeMin: stu, QCRequired: qc,
	}
	if err := a.Store.AddRoutingOperation(ctx, tx, op); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "routing_op", op.ID, actorSub, "ADD", histPayload(map[string]any{"routing_id": routingID.String()})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return op, nil
}

func (a *App) SubmitRouting(ctx context.Context, tenant, actorSub, bearer string, routingID, sedDocumentTypeID uuid.UUID, title string) error {
	if bearer == "" {
		return fmt.Errorf("%w: Authorization для sed", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	r, err := a.Store.GetRouting(ctx, tx, tenant, routingID)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrNotFound
	}
	if r.Status != "DRAFT" {
		return ErrWrongState
	}
	ops, err := a.Store.ListRoutingOperations(ctx, tx, routingID)
	if err != nil {
		return err
	}
	if len(ops) == 0 {
		return fmt.Errorf("%w: нет операций маршрута", ErrValidation)
	}
	payload, _ := json.Marshal(map[string]string{
		"kind":       "routing",
		"routing_id": routingID.String(),
	})
	doc, err := a.SED.CreateDocument(ctx, bearer, sedDocumentTypeID, title, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	if err := a.SED.SubmitDocument(ctx, bearer, doc.ID); err != nil {
		return fmt.Errorf("%w: sed submit: %v", ErrValidation, err)
	}
	sedID := doc.ID
	if err := a.Store.SetRoutingStatusSED(ctx, tx, tenant, routingID, "SUBMITTED", &sedID); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "routing", routingID, actorSub, "SUBMIT_SED", histPayload(map[string]any{"sed_document_id": sedID.String()})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
