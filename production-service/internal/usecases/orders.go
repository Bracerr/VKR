package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/production-service/internal/clients"
	"github.com/industrial-sed/production-service/internal/models"
)

func parseReservationRefs(raw json.RawMessage) ([]models.ReservationLine, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var refs []models.ReservationLine
	if err := json.Unmarshal(raw, &refs); err != nil {
		return nil, err
	}
	return refs, nil
}

func marshalReservationRefs(refs []models.ReservationLine) (json.RawMessage, error) {
	b, err := json.Marshal(refs)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// CreateProductionOrder создаёт заказ PLANNED.
func (a *App) CreateProductionOrder(ctx context.Context, tenant, actorSub string, code string, productID, bomID, routingID, warehouseID, defaultBinID uuid.UUID, qtyPlanned string, startPlan, finishPlan *time.Time) (*models.ProductionOrder, error) {
	qpDec, err := decimal.NewFromString(qtyPlanned)
	if err != nil || qpDec.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: qty_planned", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	bom, err := a.Store.GetBOM(ctx, tx, tenant, bomID)
	if err != nil {
		return nil, err
	}
	if bom == nil {
		return nil, ErrNotFound
	}
	rt, err := a.Store.GetRouting(ctx, tx, tenant, routingID)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, ErrNotFound
	}
	if code == "" {
		code = fmt.Sprintf("PO-%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
	}
	id := uuid.New()
	o := &models.ProductionOrder{
		ID: id, TenantCode: tenant, Code: code, ProductID: productID,
		QtyPlanned: qpDec.StringFixed(8), QtyDone: "0", QtyScrap: "0",
		Status: "PLANNED", BomID: bomID, RoutingID: routingID,
		WarehouseID: warehouseID, DefaultBinID: defaultBinID,
		Reservations: json.RawMessage(`[]`),
		StartPlan: startPlan, FinishPlan: finishPlan,
	}
	if err := a.Store.CreateProductionOrder(ctx, tx, o); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "production_order", id, actorSub, "CREATE", histPayload(map[string]any{"code": code})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return o, nil
}

func (a *App) ListProductionOrders(ctx context.Context, tenant string, status *string) ([]models.ProductionOrder, error) {
	return a.Store.ListProductionOrders(ctx, nil, tenant, status)
}

func (a *App) GetProductionOrder(ctx context.Context, tenant string, id uuid.UUID) (*models.ProductionOrder, error) {
	o, err := a.Store.GetProductionOrder(ctx, nil, tenant, id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, ErrNotFound
	}
	return o, nil
}

// GetProductionOrderDetail заказ и операции.
func (a *App) GetProductionOrderDetail(ctx context.Context, tenant string, id uuid.UUID) (*models.ProductionOrder, []models.OrderOperation, error) {
	o, err := a.Store.GetProductionOrder(ctx, nil, tenant, id)
	if err != nil {
		return nil, nil, err
	}
	if o == nil {
		return nil, nil, ErrNotFound
	}
	ops, err := a.Store.ListOrderOperations(ctx, nil, id)
	if err != nil {
		return nil, nil, err
	}
	return o, ops, nil
}

// ReleaseProductionOrder резерв материалов и snapshot операций.
func (a *App) ReleaseProductionOrder(ctx context.Context, tenant, actorSub string, orderID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	o, err := a.Store.LockProductionOrder(ctx, tx, tenant, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return ErrNotFound
	}
	if o.Status != "PLANNED" {
		return ErrWrongState
	}
	bom, err := a.Store.GetBOM(ctx, tx, tenant, o.BomID)
	if err != nil || bom == nil {
		return ErrNotFound
	}
	if bom.Status != "APPROVED" {
		return fmt.Errorf("%w: BOM не утверждён", ErrValidation)
	}
	rt, err := a.Store.GetRouting(ctx, tx, tenant, o.RoutingID)
	if err != nil || rt == nil {
		return ErrNotFound
	}
	if rt.Status != "APPROVED" {
		return fmt.Errorf("%w: маршрут не утверждён", ErrValidation)
	}
	lines, err := a.Store.ListBOMLines(ctx, tx, o.BomID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("%w: пустой BOM", ErrValidation)
	}
	rops, err := a.Store.ListRoutingOperations(ctx, tx, o.RoutingID)
	if err != nil {
		return err
	}
	if len(rops) == 0 {
		return fmt.Errorf("%w: пустой маршрут", ErrValidation)
	}
	qpOrder, err := decimal.NewFromString(o.QtyPlanned)
	if err != nil {
		return err
	}

	var whLines []clients.WarehousePayloadLine
	for _, ln := range lines {
		qtyNeed := bomLineQtyNeed(ln, qpOrder)
		bin := o.DefaultBinID.String()
		reason := "production_release"
		docRef := o.Code
		whLines = append(whLines, clients.WarehousePayloadLine{
			BinID: &bin, ProductID: ln.ComponentProductID.String(),
			Qty: qtyNeed.StringFixed(8), Reason: reason, DocRef: docRef,
		})
	}
	payload := &clients.WarehousePayload{
		WarehouseID:  o.WarehouseID.String(),
		DefaultBinID: func() *string { s := o.DefaultBinID.String(); return &s }(),
		Lines:        whLines,
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	ids, err := a.WH.CreateReservations(ctx, tenant, actorSub, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWarehouse, err)
	}
	if len(ids) != len(lines) {
		for _, rid := range ids {
			_ = a.WH.ReleaseReservation(ctx, tenant, rid)
		}
		return fmt.Errorf("%w: количество резервов", ErrValidation)
	}

	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		for _, rid := range ids {
			_ = a.WH.ReleaseReservation(ctx, tenant, rid)
		}
		return err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	o2, err := a.Store.LockProductionOrder(ctx, tx2, tenant, orderID)
	if err != nil || o2 == nil {
		for _, rid := range ids {
			_ = a.WH.ReleaseReservation(ctx, tenant, rid)
		}
		return ErrNotFound
	}
	if o2.Status != "PLANNED" {
		for _, rid := range ids {
			_ = a.WH.ReleaseReservation(ctx, tenant, rid)
		}
		return ErrConflict
	}
	var refs []models.ReservationLine
	for i := range lines {
		refs = append(refs, models.ReservationLine{BomLineID: lines[i].ID, ReservationID: ids[i]})
	}
	rawRefs, err := marshalReservationRefs(refs)
	if err != nil {
		for _, rid := range ids {
			_ = a.WH.ReleaseReservation(ctx, tenant, rid)
		}
		return err
	}
	for _, ro := range rops {
		op := &models.OrderOperation{
			ID: uuid.New(), OrderID: o2.ID, OpNo: ro.OpNo, WorkcenterID: ro.WorkcenterID,
			Name: ro.Name, QtyPlanned: o2.QtyPlanned, QtyGood: "0", QtyScrap: "0", Status: "PENDING",
		}
		if err := a.Store.InsertOrderOperation(ctx, tx2, op); err != nil {
			for _, rid := range ids {
				_ = a.WH.ReleaseReservation(ctx, tenant, rid)
			}
			return err
		}
	}
	if err := a.Store.UpdateProductionOrderReservations(ctx, tx2, tenant, orderID, rawRefs, "RELEASED"); err != nil {
		for _, rid := range ids {
			_ = a.WH.ReleaseReservation(ctx, tenant, rid)
		}
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx2, tenant, "production_order", orderID, actorSub, "RELEASE", histPayload(map[string]any{"reservations": len(refs)})); err != nil {
		return err
	}
	return tx2.Commit(ctx)
}

func bomLineQtyNeed(ln models.BOMLine, qtyPlannedOrder decimal.Decimal) decimal.Decimal {
	qp, _ := decimal.NewFromString(ln.QtyPer)
	scr, _ := decimal.NewFromString(ln.ScrapPct)
	mul := decimal.NewFromInt(1).Add(scr.Div(decimal.NewFromInt(100)))
	return qp.Mul(qtyPlannedOrder).Mul(mul)
}

// CancelProductionOrder отмена и снятие резервов.
func (a *App) CancelProductionOrder(ctx context.Context, tenant, actorSub string, orderID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	o, err := a.Store.LockProductionOrder(ctx, tx, tenant, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return ErrNotFound
	}
	switch o.Status {
	case "PLANNED", "RELEASED", "IN_PROGRESS":
	default:
		return ErrWrongState
	}
	if o.Status == "PLANNED" {
		if err := a.Store.UpdateProductionOrderStatus(ctx, tx, tenant, orderID, "CANCELLED"); err != nil {
			return err
		}
		if err := a.Store.InsertHistory(ctx, tx, tenant, "production_order", orderID, actorSub, "CANCEL", nil); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	refs, err := parseReservationRefs(o.Reservations)
	if err != nil {
		return err
	}
	if err := a.Store.UpdateProductionOrderStatus(ctx, tx, tenant, orderID, "CANCELLED"); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "production_order", orderID, actorSub, "CANCEL", histPayload(map[string]any{"released_reservations": len(refs)})); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	for _, r := range refs {
		if err := a.WH.ReleaseReservation(ctx, tenant, r.ReservationID); err != nil {
			return fmt.Errorf("%w: release %s: %v", ErrWarehouse, r.ReservationID, err)
		}
	}
	return nil
}

// CompleteProductionOrder приход ГП и закрытие.
func (a *App) CompleteProductionOrder(ctx context.Context, tenant, actorSub string, orderID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	o, err := a.Store.LockProductionOrder(ctx, tx, tenant, orderID)
	if err != nil {
		return err
	}
	if o == nil {
		return ErrNotFound
	}
	if o.Status != "IN_PROGRESS" && o.Status != "RELEASED" {
		return ErrWrongState
	}
	ops, err := a.Store.ListOrderOperations(ctx, tx, orderID)
	if err != nil {
		return err
	}
	for _, op := range ops {
		if op.Status != "DONE" {
			return fmt.Errorf("%w: не все операции завершены", ErrValidation)
		}
	}
	qtyDone, err := decimal.NewFromString(o.QtyDone)
	if err != nil || qtyDone.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: нет учтённой годной продукции", ErrValidation)
	}
	bin := o.DefaultBinID.String()
	payload := &clients.WarehousePayload{
		WarehouseID:  o.WarehouseID.String(),
		DefaultBinID: &bin,
		Lines: []clients.WarehousePayloadLine{
			{ProductID: o.ProductID.String(), Qty: qtyDone.StringFixed(8), Reason: "production_output", DocRef: o.Code},
		},
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	docID, err := a.WH.Receipt(ctx, tenant, actorSub, payload)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWarehouse, err)
	}
	tx3, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx3.Rollback(ctx) }()
	now := time.Now().UTC()
	if err := a.Store.SetProductionOrderReceipt(ctx, tx3, tenant, orderID, docID, "COMPLETED", now); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx3, tenant, "production_order", orderID, actorSub, "COMPLETE", histPayload(map[string]any{"warehouse_document_id": docID.String()})); err != nil {
		return err
	}
	return tx3.Commit(ctx)
}
