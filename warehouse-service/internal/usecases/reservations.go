package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// ReservationIn параметры резерва.
type ReservationIn struct {
	WarehouseID uuid.UUID
	BinID       uuid.UUID
	ProductID   uuid.UUID
	BatchID     *uuid.UUID
	SerialNo    *string
	Qty         decimal.Decimal
	Reason      string
	DocRef      string
	ExpiresAt   *time.Time
}

// CreateReservation резервирует количество (явная партия или NONE; серийник — по serial_no).
func (u *UC) CreateReservation(ctx context.Context, tenant, user string, in ReservationIn) (uuid.UUID, error) {
	if in.Qty.LessThanOrEqual(decimal.Zero) {
		return uuid.Nil, fmt.Errorf("%w: qty", ErrValidation)
	}
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := u.ensureWhBinTenant(ctx, tx, tenant, in.WarehouseID, in.BinID); err != nil {
		return uuid.Nil, err
	}

	p, err := u.Store.GetProductByID(ctx, tx, tenant, in.ProductID)
	if err != nil || p == nil {
		return uuid.Nil, ErrNotFound
	}

	var serialID *uuid.UUID
	batchID := models.NilBatchID

	switch {
	case in.SerialNo != nil && *in.SerialNo != "":
		if p.TrackingMode != models.TrackingSerial && p.TrackingMode != models.TrackingBatchAndSerial {
			return uuid.Nil, fmt.Errorf("%w: серийный резерв только для SERIAL/BATCH_AND_SERIAL", ErrValidation)
		}
		if !in.Qty.Equal(decimal.NewFromInt(1)) {
			return uuid.Nil, fmt.Errorf("%w: для серийника qty=1", ErrValidation)
		}
		s, err := u.Store.GetSerialByProductNo(ctx, tx, in.ProductID, *in.SerialNo)
		if err != nil || s == nil || s.TenantCode != tenant {
			return uuid.Nil, ErrNotFound
		}
		if s.Status != "IN_STOCK" || s.WarehouseID == nil || *s.WarehouseID != in.WarehouseID ||
			s.BinID == nil || *s.BinID != in.BinID {
			return uuid.Nil, fmt.Errorf("%w: серийник не на указанной локации", ErrValidation)
		}
		sid := s.ID
		serialID = &sid
		if s.BatchID != nil {
			batchID = *s.BatchID
		}
		bal, err := u.Store.LockBalance(ctx, tx, in.WarehouseID, in.BinID, in.ProductID, batchID)
		if err != nil {
			return uuid.Nil, err
		}
		if bal == nil || bal.Quantity.Sub(bal.ReservedQty).LessThan(decimal.NewFromInt(1)) {
			return uuid.Nil, ErrInsufficient
		}
		if err := u.Store.UpdateBalanceDeltaExisting(ctx, tx, in.WarehouseID, in.BinID, in.ProductID, batchID, decimal.Zero, decimal.NewFromInt(1), decimal.Zero); err != nil {
			return uuid.Nil, err
		}
		if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, s.ID, s.WarehouseID, s.BinID, "RESERVED", nil); err != nil {
			return uuid.Nil, err
		}

	default:
		if p.TrackingMode == models.TrackingSerial {
			return uuid.Nil, fmt.Errorf("%w: укажите serial_no", ErrValidation)
		}
		if p.TrackingMode == models.TrackingNone {
			if in.BatchID != nil {
				return uuid.Nil, fmt.Errorf("%w: batch_id не для NONE", ErrValidation)
			}
			batchID = models.NilBatchID
		} else {
			if in.BatchID == nil {
				return uuid.Nil, fmt.Errorf("%w: batch_id обязателен", ErrValidation)
			}
			batchID = *in.BatchID
		}

		bal, err := u.Store.LockBalance(ctx, tx, in.WarehouseID, in.BinID, in.ProductID, batchID)
		if err != nil {
			return uuid.Nil, err
		}
		if bal == nil {
			return uuid.Nil, ErrInsufficient
		}
		avail := bal.Quantity.Sub(bal.ReservedQty)
		if avail.LessThan(in.Qty) {
			return uuid.Nil, ErrInsufficient
		}
		if err := u.Store.UpdateBalanceDeltaExisting(ctx, tx, in.WarehouseID, in.BinID, in.ProductID, batchID, decimal.Zero, in.Qty, decimal.Zero); err != nil {
			return uuid.Nil, err
		}
	}

	r := &models.Reservation{
		ID:          uuid.New(),
		TenantCode:  tenant,
		Status:      "ACTIVE",
		WarehouseID: in.WarehouseID,
		BinID:       &in.BinID,
		ProductID:   in.ProductID,
		BatchID:     batchID,
		SerialID:    serialID,
		Qty:         in.Qty,
		Reason:      in.Reason,
		DocRef:      in.DocRef,
		ExpiresAt:   in.ExpiresAt,
		CreatedBy:   user,
	}
	if err := u.Store.CreateReservation(ctx, tx, r); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return r.ID, nil
}

// ReleaseReservation снимает резерв.
func (u *UC) ReleaseReservation(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	r, err := u.Store.GetReservation(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrNotFound
	}
	if r.Status != "ACTIVE" {
		return fmt.Errorf("%w: резерв не активен", ErrValidation)
	}
	if r.BinID == nil {
		return fmt.Errorf("%w: у резерва нет bin_id", ErrValidation)
	}

	if err := u.Store.UpsertBalanceDelta(ctx, tx, r.WarehouseID, *r.BinID, r.ProductID, r.BatchID, decimal.Zero, r.Qty.Neg(), decimal.Zero); err != nil {
		return err
	}
	if r.SerialID != nil {
		s, err := u.Store.GetSerialByID(ctx, tx, tenant, *r.SerialID)
		if err != nil || s == nil {
			return ErrNotFound
		}
		w := r.WarehouseID
		b := *r.BinID
		if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, *r.SerialID, &w, &b, "IN_STOCK", nil); err != nil {
			return err
		}
	}
	if err := u.Store.UpdateReservationStatus(ctx, tx, id, "RELEASED"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ConsumeReservation списывает зарезервированный остаток (движение RESERVE_CONSUMED).
func (u *UC) ConsumeReservation(ctx context.Context, tenant, user string, id uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	r, err := u.Store.GetReservation(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if r == nil {
		return ErrNotFound
	}
	if r.Status != "ACTIVE" {
		return fmt.Errorf("%w: резерв не активен", ErrValidation)
	}
	if r.BinID == nil {
		return fmt.Errorf("%w: у резерва нет bin_id", ErrValidation)
	}

	bal, err := u.Store.LockBalance(ctx, tx, r.WarehouseID, *r.BinID, r.ProductID, r.BatchID)
	if err != nil {
		return err
	}
	if bal == nil {
		return ErrInsufficient
	}
	if bal.ReservedQty.LessThan(r.Qty) {
		return ErrInsufficient
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
		ID: uuid.New(), TenantCode: tenant, MovementType: "RESERVE_CONSUMED", DocumentID: nil,
		WarehouseID: r.WarehouseID, BinID: *r.BinID, ProductID: r.ProductID, BatchID: r.BatchID, SerialID: r.SerialID,
		Qty: negQty, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
		return err
	}
	if err := u.Store.UpsertBalanceDelta(ctx, tx, r.WarehouseID, *r.BinID, r.ProductID, r.BatchID, negQty, negRes, val); err != nil {
		return err
	}
	if r.SerialID != nil {
		if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, *r.SerialID, nil, nil, "ISSUED", &mov.ID); err != nil {
			return err
		}
	}
	if err := u.Store.UpdateReservationStatus(ctx, tx, id, "CONSUMED"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ListReservations список резервов.
func (u *UC) ListReservations(ctx context.Context, tenant string, status *string, whID, productID *uuid.UUID) ([]models.Reservation, error) {
	return u.Store.ListReservations(ctx, nil, tenant, status, whID, productID)
}

// GetReservation возвращает резерв.
func (u *UC) GetReservation(ctx context.Context, tenant string, id uuid.UUID) (*models.Reservation, error) {
	return u.Store.GetReservation(ctx, nil, tenant, id)
}

// ProcessExpiredReservations снимает просроченные резервы (для фоновой задачи).
func (u *UC) ProcessExpiredReservations(ctx context.Context, now time.Time) (int, error) {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	list, err := u.Store.ListActiveExpiredReservations(ctx, tx, now)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range list {
		if r.BinID == nil {
			continue
		}
		if err := u.Store.UpsertBalanceDelta(ctx, tx, r.WarehouseID, *r.BinID, r.ProductID, r.BatchID, decimal.Zero, r.Qty.Neg(), decimal.Zero); err != nil {
			return n, err
		}
		if r.SerialID != nil {
			w := r.WarehouseID
			b := *r.BinID
			if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, *r.SerialID, &w, &b, "IN_STOCK", nil); err != nil {
				return n, err
			}
		}
		if err := u.Store.UpdateReservationStatus(ctx, tx, r.ID, "EXPIRED"); err != nil {
			return n, err
		}
		n++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return n, nil
}
