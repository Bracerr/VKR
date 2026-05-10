package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
	"github.com/industrial-sed/warehouse-service/internal/repositories"
)

// UC складские сценарии.
type UC struct {
	Store           *repositories.Store
	DefaultCurrency string
}

// ReceiptLineIn строка прихода.
type ReceiptLineIn struct {
	ProductID       uuid.UUID
	Qty             decimal.Decimal
	Series          *string
	ManufacturedAt  *time.Time
	ExpiresAt       *time.Time
	UnitCost        *decimal.Decimal
	Currency        *string
	SerialNumbers   []string
}

// Receipt приход.
func (u *UC) Receipt(ctx context.Context, tenant, user string, whID, binID uuid.UUID, lines []ReceiptLineIn) (docID uuid.UUID, err error) {
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
		DocType:         "RECEIPT",
		Number:          "RCP-" + uuid.NewString()[:8],
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
			return uuid.Nil, fmt.Errorf("%w: product", ErrNotFound)
		}
		if err := u.checkBinCapacity(ctx, tx, tenant, binID, ln.Qty); err != nil {
			return uuid.Nil, err
		}

		cur := ln.Currency
		if cur == nil {
			c := p.DefaultCurrency
			cur = &c
		}
		batchID := models.NilBatchID
		switch p.TrackingMode {
		case models.TrackingNone:
			val := ReceiptLineValue(ln.Qty, ln.UnitCost)
			mov := &models.StockMovement{
				ID: uuid.New(), TenantCode: tenant, MovementType: "RECEIPT", DocumentID: &doc.ID,
				WarehouseID: whID, BinID: binID, ProductID: ln.ProductID, BatchID: models.NilBatchID,
				Qty: ln.Qty, UnitCost: ln.UnitCost, Value: &val, Currency: cur, PostedAt: time.Now().UTC(), PostedBy: user,
			}
			if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
				return uuid.Nil, err
			}
			if err := u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, ln.ProductID, models.NilBatchID, ln.Qty, decimal.Zero, val); err != nil {
				return uuid.Nil, err
			}
			continue
		case models.TrackingBatch, models.TrackingBatchAndSerial:
			series := "DEFAULT"
			if ln.Series != nil && *ln.Series != "" {
				series = *ln.Series
			}
			b, err := u.Store.GetBatchByProductSeries(ctx, tx, ln.ProductID, series)
			if err != nil {
				return uuid.Nil, err
			}
			if b == nil {
				b = &models.Batch{
					ID:         uuid.New(),
					TenantCode: tenant,
					ProductID:  ln.ProductID,
					Series:     series,
					UnitCost:   ln.UnitCost,
					Currency:   cur,
				}
				if ln.ManufacturedAt != nil {
					t := *ln.ManufacturedAt
					b.ManufacturedAt = &t
				}
				if ln.ExpiresAt != nil {
					t := *ln.ExpiresAt
					b.ExpiresAt = &t
				}
				if err := u.Store.CreateBatch(ctx, tx, b); err != nil {
					return uuid.Nil, err
				}
				batchID = b.ID
			} else {
				batchID = b.ID
			}
		case models.TrackingSerial:
			if len(ln.SerialNumbers) == 0 {
				return uuid.Nil, fmt.Errorf("%w: serial_numbers required", ErrValidation)
			}
			q := ln.Qty.IntPart()
			if !ln.Qty.Equal(decimal.NewFromInt(q)) || int64(len(ln.SerialNumbers)) != q {
				return uuid.Nil, fmt.Errorf("%w: qty must match serial count", ErrValidation)
			}
			series := "SN-" + uuid.NewString()[:8]
			if ln.Series != nil && *ln.Series != "" {
				series = *ln.Series
			}
			b, _ := u.Store.GetBatchByProductSeries(ctx, tx, ln.ProductID, series)
			if b == nil {
				b = &models.Batch{ID: uuid.New(), TenantCode: tenant, ProductID: ln.ProductID, Series: series, UnitCost: ln.UnitCost, Currency: cur}
				if err := u.Store.CreateBatch(ctx, tx, b); err != nil {
					return uuid.Nil, err
				}
			}
			batchID = b.ID
			one := decimal.NewFromInt(1)
			lineVal := ReceiptLineValue(one, ln.UnitCost)
			for _, snStr := range ln.SerialNumbers {
				exists, _ := u.Store.GetSerialByProductNo(ctx, tx, ln.ProductID, snStr)
				if exists != nil {
					return uuid.Nil, fmt.Errorf("%w: serial %s", ErrConflict, snStr)
				}
				sid := uuid.New()
				s := &models.SerialNumber{
					ID: sid, TenantCode: tenant, ProductID: ln.ProductID, BatchID: &batchID, SerialNo: snStr,
					Status: "IN_STOCK", WarehouseID: &whID, BinID: &binID, UnitCost: ln.UnitCost,
				}
				if err := u.Store.CreateSerial(ctx, tx, s); err != nil {
					return uuid.Nil, err
				}
				v := lineVal
				mov := &models.StockMovement{
					ID: uuid.New(), TenantCode: tenant, MovementType: "RECEIPT", DocumentID: &doc.ID,
					WarehouseID: whID, BinID: binID, ProductID: ln.ProductID, BatchID: batchID, SerialID: &sid,
					Qty: one, UnitCost: ln.UnitCost, Value: &v, Currency: cur, PostedAt: time.Now().UTC(), PostedBy: user,
				}
				if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
					return uuid.Nil, err
				}
				if err := u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, ln.ProductID, batchID, one, decimal.Zero, lineVal); err != nil {
					return uuid.Nil, err
				}
			}
			continue
		default:
			return uuid.Nil, ErrValidation
		}

		val := ReceiptLineValue(ln.Qty, ln.UnitCost)
		mov := &models.StockMovement{
			ID:           uuid.New(),
			TenantCode:   tenant,
			MovementType: "RECEIPT",
			DocumentID:   &doc.ID,
			WarehouseID:  whID,
			BinID:        binID,
			ProductID:    ln.ProductID,
			BatchID:      batchID,
			Qty:          ln.Qty,
			UnitCost:     ln.UnitCost,
			Value:        &val,
			Currency:     cur,
			PostedAt:     time.Now().UTC(),
			PostedBy:     user,
		}
		if err := u.Store.InsertMovement(ctx, tx, mov); err != nil {
			return uuid.Nil, err
		}
		if err := u.Store.UpsertBalanceDelta(ctx, tx, whID, binID, ln.ProductID, batchID, ln.Qty, decimal.Zero, val); err != nil {
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return doc.ID, nil
}

func (u *UC) ensureWhBinTenant(ctx context.Context, tx pgx.Tx, tenant string, whID, binID uuid.UUID) (*models.Warehouse, error) {
	w, err := u.Store.GetWarehouseByID(ctx, tx, tenant, whID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, ErrNotFound
	}
	b, err := u.Store.GetBinByID(ctx, tx, tenant, binID)
	if err != nil {
		return nil, err
	}
	if b == nil || b.WarehouseID != whID {
		return nil, ErrNotFound
	}
	return w, nil
}

func (u *UC) checkBinCapacity(ctx context.Context, tx pgx.Tx, tenant string, binID uuid.UUID, addQty decimal.Decimal) error {
	bin, err := u.Store.GetBinByID(ctx, tx, tenant, binID)
	if err != nil || bin == nil {
		return ErrNotFound
	}
	if bin.CapacityQty == nil {
		return nil
	}
	sum, err := u.Store.SumBinQuantity(ctx, tx, binID)
	if err != nil {
		return err
	}
	if sum.Add(addQty).GreaterThan(*bin.CapacityQty) {
		return ErrCapacityExceeded
	}
	return nil
}
