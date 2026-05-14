package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// IssueIn расход.
type IssueIn struct {
	ProductID     uuid.UUID
	Qty           decimal.Decimal
	BatchID       *uuid.UUID
	SerialNumbers []string
}

// Issue списание (FEFO или явная партия / серийники).
func (u *UC) Issue(ctx context.Context, tenant, user string, whID, binID uuid.UUID, lines []IssueIn) (uuid.UUID, error) {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := u.ensureWhBinTenant(ctx, tx, tenant, whID, binID); err != nil {
		return uuid.Nil, err
	}
	doc := &models.Document{
		ID:              uuid.New(),
		TenantCode:      tenant,
		DocType:         "ISSUE",
		Number:          "ISS-" + uuid.NewString()[:8],
		Status:          "POSTED",
		WarehouseFromID: &whID,
		CreatedBy:       user,
	}
	if err := u.Store.CreateDocument(ctx, tx, doc); err != nil {
		return uuid.Nil, err
	}

	for _, ln := range lines {
		p, err := u.Store.GetProductByID(ctx, tx, tenant, ln.ProductID)
		if err != nil || p == nil {
			return uuid.Nil, ErrNotFound
		}
		cur := p.DefaultCurrency

		serialMode := p.TrackingMode == models.TrackingSerial ||
			(p.TrackingMode == models.TrackingBatchAndSerial && len(ln.SerialNumbers) > 0)
		if serialMode {
			if len(ln.SerialNumbers) == 0 {
				return uuid.Nil, fmt.Errorf("%w: serial_numbers", ErrValidation)
			}
			if err := u.issueSerialUnits(ctx, tx, tenant, user, doc.ID, whID, binID, ln.ProductID, ln.SerialNumbers, cur); err != nil {
				return uuid.Nil, err
			}
			continue
		}

		if p.TrackingMode == models.TrackingNone {
			if ln.BatchID != nil {
				return uuid.Nil, fmt.Errorf("%w: batch_id не применим к NONE", ErrValidation)
			}
			if len(ln.SerialNumbers) > 0 {
				return uuid.Nil, fmt.Errorf("%w: serial_numbers не применимы к NONE", ErrValidation)
			}
			if err := u.issueFromBatch(ctx, tx, tenant, user, doc.ID, whID, binID, ln.ProductID, models.NilBatchID, ln.Qty, cur); err != nil {
				return uuid.Nil, err
			}
			continue
		}

		// BATCH или BATCH_AND_SERIAL без перечисления серийников — списание по партиям (FEFO / явная партия).
		if len(ln.SerialNumbers) > 0 {
			return uuid.Nil, fmt.Errorf("%w: serial_numbers только для SERIAL/BATCH_AND_SERIAL", ErrValidation)
		}
		need := ln.Qty
		if ln.BatchID != nil {
			bid := *ln.BatchID
			if err := u.issueFromBatch(ctx, tx, tenant, user, doc.ID, whID, binID, ln.ProductID, bid, need, cur); err != nil {
				return uuid.Nil, err
			}
			continue
		}
		rows, err := u.Store.ListBatchesForFEFO(ctx, tx, tenant, whID, binID, ln.ProductID)
		if err != nil {
			return uuid.Nil, err
		}
		for _, row := range rows {
			if need.IsZero() {
				break
			}
			take := decimal.Min(row.QtyAvail, need)
			if take.IsZero() {
				continue
			}
			if err := u.issueFromBatch(ctx, tx, tenant, user, doc.ID, whID, binID, ln.ProductID, row.BatchID, take, cur); err != nil {
				return uuid.Nil, err
			}
			need = need.Sub(take)
		}
		if need.GreaterThan(decimal.Zero) {
			return uuid.Nil, ErrInsufficient
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	u.emitTraceDocumentPosted(ctx, tenant, doc.ID)
	return doc.ID, nil
}

func (u *UC) issueSerialUnits(ctx context.Context, tx pgx.Tx, tenant, user string, docID, whID, binID, productID uuid.UUID, serialNos []string, cur string) error {
	for _, sn := range serialNos {
		s, err := u.Store.GetSerialByProductNo(ctx, tx, productID, sn)
		if err != nil || s == nil || s.TenantCode != tenant {
			return ErrNotFound
		}
		if s.Status != "IN_STOCK" || s.WarehouseID == nil || *s.WarehouseID != whID || s.BinID == nil || *s.BinID != binID {
			return fmt.Errorf("%w: serial not in stock at location", ErrValidation)
		}
		batchID := models.NilBatchID
		if s.BatchID != nil {
			batchID = *s.BatchID
		}
		bal, err := u.Store.LockBalance(ctx, tx, whID, binID, productID, batchID)
		if err != nil {
			return err
		}
		if bal == nil || bal.Quantity.Sub(bal.ReservedQty).LessThan(decimal.NewFromInt(1)) {
			return ErrInsufficient
		}
		neg := decimal.NewFromInt(-1)
		val := MovementValue(decimal.NewFromInt(1), bal.Quantity, bal.Value).Neg()
		mov := &models.StockMovement{
			ID: uuid.New(), TenantCode: tenant, MovementType: "ISSUE", DocumentID: &docID,
			WarehouseID: whID, BinID: binID, ProductID: productID, BatchID: batchID, SerialID: &s.ID,
			Qty: neg, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
		}
		if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
			return err
		}
		if err := u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, productID, batchID, neg, decimal.Zero, val); err != nil {
			return err
		}
		if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, s.ID, nil, nil, "ISSUED", &mov.ID); err != nil {
			return err
		}
	}
	return nil
}

func (u *UC) issueFromBatch(ctx context.Context, tx pgx.Tx, tenant, user string, docID, whID, binID, productID, batchID uuid.UUID, qty decimal.Decimal, cur string) error {
	bal, err := u.Store.LockBalance(ctx, tx, whID, binID, productID, batchID)
	if err != nil {
		return err
	}
	if bal == nil {
		return ErrInsufficient
	}
	avail := bal.Quantity.Sub(bal.ReservedQty)
	if avail.LessThan(qty) {
		return ErrInsufficient
	}
	neg := qty.Neg()
	val := MovementValue(qty, bal.Quantity, bal.Value).Neg()
	mov := &models.StockMovement{
		ID: uuid.New(), TenantCode: tenant, MovementType: "ISSUE", DocumentID: &docID,
		WarehouseID: whID, BinID: binID, ProductID: productID, BatchID: batchID,
		Qty: neg, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
		return err
	}
	return u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, productID, batchID, neg, decimal.Zero, val)
}
