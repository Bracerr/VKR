package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/production-service/internal/models"
)

// StartOperation старт операции заказа.
func (a *App) StartOperation(ctx context.Context, tenant, actorSub string, orderID, opID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	op, err := a.Store.LockOrderOperation(ctx, tx, orderID, opID)
	if err != nil {
		return err
	}
	if op == nil {
		return ErrNotFound
	}
	if op.Status != "PENDING" {
		return ErrWrongState
	}
	o, err := a.Store.LockProductionOrder(ctx, tx, tenant, orderID)
	if err != nil || o == nil {
		return ErrNotFound
	}
	if o.Status != "RELEASED" && o.Status != "IN_PROGRESS" {
		return ErrWrongState
	}
	now := time.Now().UTC()
	op.Status = "STARTED"
	op.StartedAt = &now
	if err := a.Store.UpdateOrderOperation(ctx, tx, op); err != nil {
		return err
	}
	newStatus := o.Status
	if o.Status == "RELEASED" {
		newStatus = "IN_PROGRESS"
		st := now
		if err := a.Store.UpdateProductionOrderProgress(ctx, tx, tenant, orderID, o.QtyDone, o.QtyScrap, newStatus, &st, nil); err != nil {
			return err
		}
	} else {
		if err := a.Store.UpdateProductionOrderStatus(ctx, tx, tenant, orderID, "IN_PROGRESS"); err != nil {
			return err
		}
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "order_operation", opID, actorSub, "START", nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ReportOperation учёт выработки (без consume склада — см. FinishOperation).
func (a *App) ReportOperation(ctx context.Context, tenant, actorSub string, orderID, opID uuid.UUID, qtyGood, qtyScrap string, scrapReason *string, note *string) error {
	good, err := decimal.NewFromString(qtyGood)
	if err != nil || good.IsNegative() {
		return fmt.Errorf("%w: qty_good", ErrValidation)
	}
	scrap, err := decimal.NewFromString(qtyScrap)
	if err != nil || scrap.IsNegative() {
		return fmt.Errorf("%w: qty_scrap", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	op, err := a.Store.LockOrderOperation(ctx, tx, orderID, opID)
	if err != nil {
		return err
	}
	if op == nil {
		return ErrNotFound
	}
	if op.Status == "DONE" {
		return ErrWrongState
	}
	o, err := a.Store.LockProductionOrder(ctx, tx, tenant, orderID)
	if err != nil || o == nil {
		return ErrNotFound
	}
	if o.Status != "RELEASED" && o.Status != "IN_PROGRESS" {
		return ErrWrongState
	}

	opGood, _ := decimal.NewFromString(op.QtyGood)
	opScr, _ := decimal.NewFromString(op.QtyScrap)
	op.QtyGood = opGood.Add(good).StringFixed(8)
	op.QtyScrap = opScr.Add(scrap).StringFixed(8)

	ordDone, _ := decimal.NewFromString(o.QtyDone)
	ordScr, _ := decimal.NewFromString(o.QtyScrap)
	o.QtyDone = ordDone.Add(good).StringFixed(8)
	o.QtyScrap = ordScr.Add(scrap).StringFixed(8)

	if op.Status == "PENDING" {
		now := time.Now().UTC()
		op.Status = "STARTED"
		op.StartedAt = &now
	}

	if err := a.Store.UpdateOrderOperation(ctx, tx, op); err != nil {
		return err
	}
	st := time.Now().UTC()
	if err := a.Store.UpdateProductionOrderProgress(ctx, tx, tenant, orderID, o.QtyDone, o.QtyScrap, "IN_PROGRESS", &st, nil); err != nil {
		return err
	}
	rep := &models.ProductionReport{
		ID: uuid.New(), TenantCode: tenant, OrderOperationID: opID,
		ReportedBySub: actorSub, QtyGood: good.StringFixed(8), QtyScrap: scrap.StringFixed(8),
		ScrapReasonCode: scrapReason, Note: note,
	}
	if err := a.Store.InsertProductionReport(ctx, tx, rep); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "order_operation", opID, actorSub, "REPORT", histPayload(map[string]any{"qty_good": qtyGood, "qty_scrap": qtyScrap})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// FinishOperation завершение операции и consume резервов по строкам BOM с этим op_no.
func (a *App) FinishOperation(ctx context.Context, tenant, actorSub string, orderID, opID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	op, err := a.Store.LockOrderOperation(ctx, tx, orderID, opID)
	if err != nil {
		return err
	}
	if op == nil {
		return ErrNotFound
	}
	if op.Status == "DONE" {
		return nil
	}
	o, err := a.Store.LockProductionOrder(ctx, tx, tenant, orderID)
	if err != nil || o == nil {
		return ErrNotFound
	}
	if o.Status != "RELEASED" && o.Status != "IN_PROGRESS" {
		return ErrWrongState
	}
	lines, err := a.Store.ListBOMLines(ctx, tx, o.BomID)
	if err != nil {
		return err
	}
	refs, err := parseReservationRefs(o.Reservations)
	if err != nil {
		return err
	}
	refByLine := make(map[uuid.UUID]uuid.UUID)
	for _, r := range refs {
		refByLine[r.BomLineID] = r.ReservationID
	}
	var toConsume []uuid.UUID
	for _, ln := range lines {
		if ln.OpNo != op.OpNo {
			continue
		}
		rid, ok := refByLine[ln.ID]
		if !ok {
			continue
		}
		toConsume = append(toConsume, rid)
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	for _, rid := range toConsume {
		if err := a.WH.ConsumeReservation(ctx, tenant, rid); err != nil {
			return fmt.Errorf("%w: consume %s: %v", ErrWarehouse, rid, err)
		}
	}

	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	op2, err := a.Store.LockOrderOperation(ctx, tx2, orderID, opID)
	if err != nil || op2 == nil {
		return ErrNotFound
	}
	now := time.Now().UTC()
	op2.Status = "DONE"
	op2.FinishedAt = &now
	if err := a.Store.UpdateOrderOperation(ctx, tx2, op2); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx2, tenant, "order_operation", opID, actorSub, "FINISH", histPayload(map[string]any{"consumed_reservations": len(toConsume)})); err != nil {
		return err
	}
	return tx2.Commit(ctx)
}

// CreateShiftTask создаёт сменное задание.
func (a *App) CreateShiftTask(ctx context.Context, tenant, actorSub string, orderOpID uuid.UUID, shiftDate time.Time, shiftNo int, assignee *string, qtyPlanned *string) (*models.ShiftTask, error) {
	if shiftNo < 1 || shiftNo > 3 {
		return nil, fmt.Errorf("%w: shift_no", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	op, err := a.Store.GetOrderOperationByID(ctx, tx, tenant, orderOpID)
	if err != nil {
		return nil, err
	}
	if op == nil {
		return nil, ErrNotFound
	}
	o, err := a.Store.GetProductionOrder(ctx, tx, tenant, op.OrderID)
	if err != nil || o == nil {
		return nil, ErrNotFound
	}
	t := &models.ShiftTask{
		ID: uuid.New(), TenantCode: tenant, OrderOperationID: orderOpID,
		ShiftDate: shiftDate.UTC(), ShiftNo: shiftNo, AssigneeSub: assignee, QtyPlanned: qtyPlanned,
	}
	if err := a.Store.CreateShiftTask(ctx, tx, t); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "shift_task", t.ID, actorSub, "CREATE", nil); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return t, nil
}

func (a *App) ListShiftTasks(ctx context.Context, tenant string, date *time.Time) ([]models.ShiftTask, error) {
	return a.Store.ListShiftTasks(ctx, nil, tenant, date)
}

func (a *App) MeShiftTasks(ctx context.Context, tenant, sub string) ([]models.ShiftTask, error) {
	return a.Store.ListShiftTasksForAssignee(ctx, nil, tenant, sub)
}

func (a *App) DeleteShiftTask(ctx context.Context, tenant, actorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteShiftTask(ctx, tx, tenant, id); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, tenant, "shift_task", id, actorSub, "DELETE", nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
