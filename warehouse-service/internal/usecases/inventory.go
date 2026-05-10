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

// StartInventory создаёт черновик и строки по текущим остаткам (warehouse ± bin).
func (u *UC) StartInventory(ctx context.Context, tenant, user string, whID uuid.UUID, binID *uuid.UUID) (uuid.UUID, error) {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := u.Store.GetWarehouseByID(ctx, tx, tenant, whID); err != nil {
		return uuid.Nil, err
	}
	if binID != nil {
		if _, err := u.ensureWhBinTenant(ctx, tx, tenant, whID, *binID); err != nil {
			return uuid.Nil, err
		}
	}

	doc := &models.Document{
		ID:              uuid.New(),
		TenantCode:      tenant,
		DocType:         "INVENTORY",
		Number:          "INV-" + uuid.NewString()[:8],
		Status:          "DRAFT",
		WarehouseFromID: &whID,
		CreatedBy:       user,
	}
	if err := u.Store.CreateDocument(ctx, tx, doc); err != nil {
		return uuid.Nil, err
	}

	bals, err := u.Store.ListBalances(ctx, tx, tenant, &whID, binID, nil, nil, true, nil)
	if err != nil {
		return uuid.Nil, err
	}
	for _, b := range bals {
		if err := u.Store.InsertInventoryLine(ctx, tx, doc.ID, b.WarehouseID, b.BinID, b.ProductID, b.BatchID, nil, b.Quantity, nil); err != nil {
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return doc.ID, nil
}

// SetInventoryCounted фиксирует факт по строке инвентаризации.
func (u *UC) SetInventoryCounted(ctx context.Context, tenant string, lineID uuid.UUID, counted decimal.Decimal) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.UpdateInventoryLineCounted(ctx, tx, tenant, lineID, counted); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ListInventoryLines строки документа инвентаризации.
func (u *UC) ListInventoryLines(ctx context.Context, tenant string, docID uuid.UUID) ([]struct {
	ID, WhID, BinID, ProductID, BatchID uuid.UUID
	SerialID                             *uuid.UUID
	Expected                             decimal.Decimal
	Counted                              *decimal.Decimal
}, error) {
	return u.Store.ListInventoryLines(ctx, nil, tenant, docID)
}

// PostInventory проводит инвентаризацию (корректирующие движения).
func (u *UC) PostInventory(ctx context.Context, tenant, user string, docID uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	d, err := u.Store.GetDocument(ctx, tx, tenant, docID)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	if d.DocType != "INVENTORY" || d.Status != "DRAFT" {
		return fmt.Errorf("%w: документ не в статусе DRAFT", ErrValidation)
	}

	lines, err := u.Store.ListInventoryLines(ctx, tx, tenant, docID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("%w: нет строк", ErrValidation)
	}

	for _, ln := range lines {
		if ln.Counted == nil {
			return fmt.Errorf("%w: не введён факт по всем строкам", ErrValidation)
		}
		diff := ln.Counted.Sub(ln.Expected)
		if diff.IsZero() {
			continue
		}

		p, err := u.Store.GetProductByID(ctx, tx, tenant, ln.ProductID)
		if err != nil || p == nil {
			return ErrNotFound
		}
		cur := p.DefaultCurrency

		if ln.SerialID != nil {
			if !diff.Equal(decimal.NewFromInt(1)) && !diff.Equal(decimal.NewFromInt(-1)) {
				return fmt.Errorf("%w: по серийной строке допустим только пересчёт ±1", ErrValidation)
			}
			if diff.IsPositive() {
				return fmt.Errorf("%w: излишек по серийнику не поддерживается", ErrValidation)
			}
			if err := u.inventorySerialDecrease(ctx, tx, tenant, user, docID, ln.WhID, ln.BinID, ln.ProductID, *ln.SerialID, cur); err != nil {
				return err
			}
			continue
		}

		if diff.IsPositive() {
			unitCost := p.StandardCost
			val := ReceiptLineValue(diff, unitCost)
			mov := &models.StockMovement{
				ID: uuid.New(), TenantCode: tenant, MovementType: "INVENTORY_ADJUST", DocumentID: &docID,
				WarehouseID: ln.WhID, BinID: ln.BinID, ProductID: ln.ProductID, BatchID: ln.BatchID,
				Qty: diff, UnitCost: unitCost, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
			}
			if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
				return err
			}
			if err := u.Store.UpsertBalanceDelta(ctx, tx, ln.WhID, ln.BinID, ln.ProductID, ln.BatchID, diff, decimal.Zero, val); err != nil {
				return err
			}
			continue
		}

		qtyDown := diff.Neg()
		if err := u.inventoryDecreaseBatch(ctx, tx, tenant, user, docID, ln.WhID, ln.BinID, ln.ProductID, ln.BatchID, qtyDown, cur); err != nil {
			return err
		}
	}

	if err := u.Store.UpdateDocumentStatus(ctx, tx, tenant, docID, "POSTED"); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) inventoryDecreaseBatch(ctx context.Context, tx pgx.Tx, tenant, user string, docID, whID, binID, productID, batchID uuid.UUID, qty decimal.Decimal, cur string) error {
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
		ID: uuid.New(), TenantCode: tenant, MovementType: "INVENTORY_ADJUST", DocumentID: &docID,
		WarehouseID: whID, BinID: binID, ProductID: productID, BatchID: batchID,
		Qty: neg, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
		return err
	}
	return u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, productID, batchID, neg, decimal.Zero, val)
}

func (u *UC) inventorySerialDecrease(ctx context.Context, tx pgx.Tx, tenant, user string, docID, whID, binID, productID, serialID uuid.UUID, cur string) error {
	s, err := u.Store.GetSerialByID(ctx, tx, tenant, serialID)
	if err != nil || s == nil || s.ProductID != productID {
		return ErrNotFound
	}
	if s.Status != "IN_STOCK" {
		return fmt.Errorf("%w: серийник не IN_STOCK", ErrValidation)
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
		ID: uuid.New(), TenantCode: tenant, MovementType: "INVENTORY_ADJUST", DocumentID: &docID,
		WarehouseID: whID, BinID: binID, ProductID: productID, BatchID: batchID, SerialID: &serialID,
		Qty: neg, Value: &val, Currency: &cur, PostedAt: time.Now().UTC(), PostedBy: user,
	}
	if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
		return err
	}
	if err := u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, productID, batchID, neg, decimal.Zero, val); err != nil {
		return err
	}
	return u.Store.UpdateSerialLocationAndStatus(ctx, tx, serialID, nil, nil, "SCRAPPED", &mov.ID)
}
