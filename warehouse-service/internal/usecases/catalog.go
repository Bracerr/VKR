package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// --- Products ---

func (u *UC) ListProducts(ctx context.Context, tenant string) ([]models.Product, error) {
	return u.Store.ListProducts(ctx, nil, tenant)
}

func (u *UC) GetProduct(ctx context.Context, tenant string, id uuid.UUID) (*models.Product, error) {
	return u.Store.GetProductByID(ctx, nil, tenant, id)
}

func (u *UC) CreateProduct(ctx context.Context, p *models.Product) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	ok, err := u.Store.ProductExistsSKU(ctx, tx, p.TenantCode, p.SKU)
	if err != nil {
		return err
	}
	if ok {
		return ErrConflict
	}
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if err := u.Store.CreateProduct(ctx, tx, p); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) UpdateProduct(ctx context.Context, p *models.Product) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	old, err := u.Store.GetProductByID(ctx, tx, p.TenantCode, p.ID)
	if err != nil {
		return err
	}
	if old == nil {
		return ErrNotFound
	}
	if err := u.Store.UpdateProduct(ctx, tx, p); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) DeleteProduct(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.DeleteProduct(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

// --- Warehouses ---

func (u *UC) ListWarehouses(ctx context.Context, tenant string) ([]models.Warehouse, error) {
	return u.Store.ListWarehouses(ctx, nil, tenant)
}

func (u *UC) CreateWarehouse(ctx context.Context, w *models.Warehouse) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if err := u.Store.CreateWarehouse(ctx, tx, w); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) UpdateWarehouse(ctx context.Context, w *models.Warehouse) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.UpdateWarehouse(ctx, tx, w); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) DeleteWarehouse(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.DeleteWarehouse(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

// --- Bins ---

func (u *UC) ListBins(ctx context.Context, tenant string, whID uuid.UUID) ([]models.Bin, error) {
	return u.Store.ListBinsByWarehouse(ctx, nil, tenant, whID)
}

func (u *UC) CreateBin(ctx context.Context, b *models.Bin) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := u.Store.GetWarehouseByID(ctx, tx, b.TenantCode, b.WarehouseID); err != nil {
		return err
	}
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if err := u.Store.CreateBin(ctx, tx, b); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) UpdateBin(ctx context.Context, b *models.Bin) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.UpdateBin(ctx, tx, b); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) DeleteBin(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.DeleteBin(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

// --- Batches (read-only справочник) ---

// ListSerials серийные номера.
func (u *UC) ListSerials(ctx context.Context, tenant string, productID *uuid.UUID, status *string, whID *uuid.UUID) ([]models.SerialNumber, error) {
	return u.Store.ListSerials(ctx, nil, tenant, productID, status, whID, nil, nil)
}

// SerialMovementHistory журнал по серийнику.
func (u *UC) SerialMovementHistory(ctx context.Context, serialID uuid.UUID) ([]models.StockMovement, error) {
	return u.Store.MovementHistoryForSerial(ctx, nil, serialID)
}

func (u *UC) GetBatch(ctx context.Context, tenant string, id uuid.UUID) (*models.Batch, error) {
	b, err := u.Store.GetBatchByID(ctx, nil, tenant, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, ErrNotFound
	}
	return b, nil
}

// --- Prices ---

func (u *UC) ListPrices(ctx context.Context, tenant string, productID uuid.UUID) ([]models.ProductPrice, error) {
	return u.Store.ListPrices(ctx, nil, tenant, productID)
}

func (u *UC) CreatePrice(ctx context.Context, pr *models.ProductPrice) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	overlap, err := u.Store.OverlappingPrice(ctx, tx, pr.TenantCode, pr.ProductID, pr.PriceType, pr.ValidFrom, pr.ValidTo, nil)
	if err != nil {
		return err
	}
	if overlap {
		return fmt.Errorf("%w: пересечение периода цены", ErrConflict)
	}
	if pr.ID == uuid.Nil {
		pr.ID = uuid.New()
	}
	if err := u.Store.CreatePrice(ctx, tx, pr); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *UC) DeletePrice(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := u.Store.DeletePrice(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}
