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

// TransferLine перемещение партии/серийников между локациями.
type TransferLine struct {
	ProductID     uuid.UUID
	Qty           decimal.Decimal
	BatchID       *uuid.UUID
	SerialNumbers []string
}

// Transfer между складами (и ячейками).
func (u *UC) Transfer(ctx context.Context, tenant, user string, whFrom, binFrom, whTo, binTo uuid.UUID, lines []TransferLine) (uuid.UUID, error) {
	return u.moveStock(ctx, tenant, user, whFrom, binFrom, whTo, binTo, lines, false)
}

// Relocate между ячейками одного склада.
func (u *UC) Relocate(ctx context.Context, tenant, user string, whID, binFrom, binTo uuid.UUID, lines []TransferLine) (uuid.UUID, error) {
	if binFrom == binTo {
		return uuid.Nil, fmt.Errorf("%w: одинаковые ячейки", ErrValidation)
	}
	return u.moveStock(ctx, tenant, user, whID, binFrom, whID, binTo, lines, true)
}

func (u *UC) moveStock(ctx context.Context, tenant, user string, whFrom, binFrom, whTo, binTo uuid.UUID, lines []TransferLine, relocate bool) (uuid.UUID, error) {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := u.ensureWhBinTenant(ctx, tx, tenant, whFrom, binFrom); err != nil {
		return uuid.Nil, err
	}
	if _, err := u.ensureWhBinTenant(ctx, tx, tenant, whTo, binTo); err != nil {
		return uuid.Nil, err
	}
	if relocate && whFrom != whTo {
		return uuid.Nil, fmt.Errorf("%w: relocate только внутри склада", ErrValidation)
	}

	docType := "TRANSFER"
	movOut := "TRANSFER_OUT"
	movIn := "TRANSFER_IN"
	if relocate {
		docType = "RELOCATE"
		movOut = "RELOCATE_OUT"
		movIn = "RELOCATE_IN"
	}

	doc := &models.Document{
		ID:              uuid.New(),
		TenantCode:      tenant,
		DocType:         docType,
		Number:          docType[:3] + "-" + uuid.NewString()[:8],
		Status:          "POSTED",
		WarehouseFromID: &whFrom,
		WarehouseToID:   &whTo,
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
			for _, sn := range ln.SerialNumbers {
				if err := u.moveSerial(ctx, tx, tenant, user, doc.ID, whFrom, binFrom, whTo, binTo, ln.ProductID, sn, cur, movOut, movIn); err != nil {
					return uuid.Nil, err
				}
			}
			continue
		}

		if p.TrackingMode == models.TrackingNone {
			if ln.BatchID != nil || len(ln.SerialNumbers) > 0 {
				return uuid.Nil, ErrValidation
			}
			if err := u.checkBinCapacity(ctx, tx, tenant, binTo, ln.Qty); err != nil {
				return uuid.Nil, err
			}
			if err := u.moveBatchQty(ctx, tx, tenant, user, doc.ID, whFrom, binFrom, whTo, binTo, ln.ProductID, models.NilBatchID, ln.Qty, cur, movOut, movIn); err != nil {
				return uuid.Nil, err
			}
			continue
		}

		if len(ln.SerialNumbers) > 0 {
			return uuid.Nil, ErrValidation
		}

		need := ln.Qty
		if ln.BatchID != nil {
			if err := u.checkBinCapacity(ctx, tx, tenant, binTo, need); err != nil {
				return uuid.Nil, err
			}
			if err := u.moveBatchQty(ctx, tx, tenant, user, doc.ID, whFrom, binFrom, whTo, binTo, ln.ProductID, *ln.BatchID, need, cur, movOut, movIn); err != nil {
				return uuid.Nil, err
			}
			continue
		}

		rows, err := u.Store.ListBatchesForFEFO(ctx, tx, tenant, whFrom, binFrom, ln.ProductID)
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
			if err := u.checkBinCapacity(ctx, tx, tenant, binTo, take); err != nil {
				return uuid.Nil, err
			}
			if err := u.moveBatchQty(ctx, tx, tenant, user, doc.ID, whFrom, binFrom, whTo, binTo, ln.ProductID, row.BatchID, take, cur, movOut, movIn); err != nil {
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
	return doc.ID, nil
}

func (u *UC) moveBatchQty(ctx context.Context, tx pgx.Tx, tenant, user string, docID, whFrom, binFrom, whTo, binTo, productID, batchID uuid.UUID, qty decimal.Decimal, cur, movOut, movIn string) error {
	bal, err := u.Store.LockBalance(ctx, tx, whFrom, binFrom, productID, batchID)
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
	val := MovementValue(qty, bal.Quantity, bal.Value)
	neg := qty.Neg()
	vNeg := val.Neg()
	out := &models.StockMovement{
		ID: uuid.New(), TenantCode: tenant, MovementType: movOut, DocumentID: &docID,
		WarehouseID: whFrom, BinID: binFrom, ProductID: productID, BatchID: batchID,
		Qty: neg, Value: &vNeg, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, out); err != nil {
		return err
	}
	if err := u.Store.UpsertBalanceDelta(ctx, tx, whFrom, binFrom, productID, batchID, neg, decimal.Zero, vNeg); err != nil {
		return err
	}
	in := &models.StockMovement{
		ID: uuid.New(), TenantCode: tenant, MovementType: movIn, DocumentID: &docID,
		WarehouseID: whTo, BinID: binTo, ProductID: productID, BatchID: batchID,
		Qty: qty, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, in); err != nil {
		return err
	}
	return u.Store.UpsertBalanceDelta(ctx, tx, whTo, binTo, productID, batchID, qty, decimal.Zero, val)
}

func (u *UC) moveSerial(ctx context.Context, tx pgx.Tx, tenant, user string, docID, whFrom, binFrom, whTo, binTo uuid.UUID, productID uuid.UUID, serialNo, cur, movOut, movIn string) error {
	s, err := u.Store.GetSerialByProductNo(ctx, tx, productID, serialNo)
	if err != nil || s == nil || s.TenantCode != tenant {
		return ErrNotFound
	}
	if s.Status != "IN_STOCK" || s.WarehouseID == nil || *s.WarehouseID != whFrom || s.BinID == nil || *s.BinID != binFrom {
		return fmt.Errorf("%w: serial not in stock at source", ErrValidation)
	}
	batchID := models.NilBatchID
	if s.BatchID != nil {
		batchID = *s.BatchID
	}
	bal, err := u.Store.LockBalance(ctx, tx, whFrom, binFrom, productID, batchID)
	if err != nil {
		return err
	}
	if bal == nil || bal.Quantity.Sub(bal.ReservedQty).LessThan(decimal.NewFromInt(1)) {
		return ErrInsufficient
	}
	one := decimal.NewFromInt(1)
	val := MovementValue(one, bal.Quantity, bal.Value)
	vNeg := val.Neg()
	neg := one.Neg()

	if err := u.checkBinCapacity(ctx, tx, tenant, binTo, one); err != nil {
		return err
	}

	out := &models.StockMovement{
		ID: uuid.New(), TenantCode: tenant, MovementType: movOut, DocumentID: &docID,
		WarehouseID: whFrom, BinID: binFrom, ProductID: productID, BatchID: batchID, SerialID: &s.ID,
		Qty: neg, Value: &vNeg, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, out); err != nil {
		return err
	}
	if err := u.Store.UpsertBalanceDelta(ctx, tx, whFrom, binFrom, productID, batchID, neg, decimal.Zero, vNeg); err != nil {
		return err
	}
	wTr := whFrom
	bTr := binFrom
	if err := u.Store.UpdateSerialLocationAndStatus(ctx, tx, s.ID, &wTr, &bTr, "IN_TRANSIT", &out.ID); err != nil {
		return err
	}

	in := &models.StockMovement{
		ID: uuid.New(), TenantCode: tenant, MovementType: movIn, DocumentID: &docID,
		WarehouseID: whTo, BinID: binTo, ProductID: productID, BatchID: batchID, SerialID: &s.ID,
		Qty: one, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, in); err != nil {
		return err
	}
	if err := u.Store.UpsertBalanceDelta(ctx, tx, whTo, binTo, productID, batchID, one, decimal.Zero, val); err != nil {
		return err
	}
	wTo := whTo
	bTo := binTo
	return u.Store.UpdateSerialLocationAndStatus(ctx, tx, s.ID, &wTo, &bTo, "IN_STOCK", &in.ID)
}
