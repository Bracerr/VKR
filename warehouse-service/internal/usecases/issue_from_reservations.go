package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// IssueFromReservations списывает по активным резервам и создаёт складской документ ISSUE.
func (u *UC) IssueFromReservations(ctx context.Context, tenant, user string, reservationIDs []uuid.UUID) (uuid.UUID, error) {
	if len(reservationIDs) == 0 {
		return uuid.Nil, fmt.Errorf("%w: reservation_ids", ErrValidation)
	}
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	doc := &models.Document{
		ID:         uuid.New(),
		TenantCode: tenant,
		DocType:    "ISSUE",
		Number:     "ISS-RSV-" + uuid.NewString()[:8],
		Status:     "POSTED",
		CreatedBy:  user,
	}
	if err := u.Store.CreateDocument(ctx, tx, doc); err != nil {
		return uuid.Nil, err
	}

	for _, rid := range reservationIDs {
		r, err := u.Store.GetReservation(ctx, tx, tenant, rid)
		if err != nil {
			return uuid.Nil, err
		}
		if r == nil {
			return uuid.Nil, ErrNotFound
		}
		if r.Status != "ACTIVE" {
			return uuid.Nil, fmt.Errorf("%w: резерв не активен", ErrValidation)
		}
		if r.BinID == nil {
			return uuid.Nil, fmt.Errorf("%w: у резерва нет bin_id", ErrValidation)
		}

		bal, err := u.Store.LockBalance(ctx, tx, r.WarehouseID, *r.BinID, r.ProductID, r.BatchID)
		if err != nil {
			return uuid.Nil, err
		}
		if bal == nil {
			return uuid.Nil, ErrInsufficient
		}
		// должно хватать и quantity, и reserved_qty
		if bal.Quantity.LessThan(r.Qty) || bal.ReservedQty.LessThan(r.Qty) {
			return uuid.Nil, ErrInsufficient
		}

		cur := ""
		p, _ := u.Store.GetProductByID(ctx, tx, tenant, r.ProductID)
		if p != nil {
			cur = p.DefaultCurrency
		}

		negQty := r.Qty.Neg()
		negRes := r.Qty.Neg()
		val := MovementValue(r.Qty, bal.Quantity, bal.Value).Neg()

		mov := &models.StockMovement{
			ID: uuid.New(), TenantCode: tenant, MovementType: "ISSUE", DocumentID: &doc.ID,
			WarehouseID: r.WarehouseID, BinID: *r.BinID, ProductID: r.ProductID, BatchID: r.BatchID, SerialID: r.SerialID,
			Qty: negQty, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
		}
		if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
			return uuid.Nil, err
		}
		// баланс точно существует (LockBalance), поэтому обновляем существующую строку
		// (INSERT ... ON CONFLICT может провалиться на CHECK до конфликта).
		if err := u.Store.UpdateBalanceDeltaExisting(ctx, tx, r.WarehouseID, *r.BinID, r.ProductID, r.BatchID, negQty, negRes, val); err != nil {
			return uuid.Nil, err
		}
		if r.SerialID != nil {
			if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, *r.SerialID, nil, nil, "ISSUED", &mov.ID); err != nil {
				return uuid.Nil, err
			}
		}
		if err := u.Store.UpdateReservationStatus(ctx, tx, rid, "CONSUMED"); err != nil {
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	u.emitTraceDocumentPosted(ctx, tenant, doc.ID)
	return doc.ID, nil
}

