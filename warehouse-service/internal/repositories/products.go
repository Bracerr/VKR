package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// CreateProduct создаёт товар.
func (s *Store) CreateProduct(ctx context.Context, tx pgx.Tx, p *models.Product) error {
	var std *string
	if p.StandardCost != nil {
		v := p.StandardCost.StringFixed(4)
		std = &v
	}
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO products (id, tenant_code, sku, name, unit, tracking_mode, has_expiration, valuation_method, default_currency, standard_cost)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, p.ID, p.TenantCode, p.SKU, p.Name, p.Unit, p.TrackingMode, p.HasExpiration, p.ValuationMethod, p.DefaultCurrency, std)
	return err
}

// GetProductByID товар по id и tenant.
func (s *Store) GetProductByID(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Product, error) {
	var p models.Product
	var std *string
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, sku, name, unit, tracking_mode, has_expiration, valuation_method, default_currency,
		       standard_cost::text, created_at, updated_at
		FROM products WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(
		&p.ID, &p.TenantCode, &p.SKU, &p.Name, &p.Unit, &p.TrackingMode, &p.HasExpiration, &p.ValuationMethod, &p.DefaultCurrency,
		&std, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.StandardCost = decPtr(std)
	return &p, nil
}

// ListProducts список товаров тенанта.
func (s *Store) ListProducts(ctx context.Context, tx pgx.Tx, tenant string) ([]models.Product, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, sku, name, unit, tracking_mode, has_expiration, valuation_method, default_currency,
		       standard_cost::text, created_at, updated_at
		FROM products WHERE tenant_code = $1 ORDER BY sku
	`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Product
	for rows.Next() {
		var p models.Product
		var std *string
		if err := rows.Scan(&p.ID, &p.TenantCode, &p.SKU, &p.Name, &p.Unit, &p.TrackingMode, &p.HasExpiration, &p.ValuationMethod, &p.DefaultCurrency,
			&std, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.StandardCost = decPtr(std)
		out = append(out, p)
	}
	return out, rows.Err()
}

// UpdateProduct обновление.
func (s *Store) UpdateProduct(ctx context.Context, tx pgx.Tx, p *models.Product) error {
	var std *string
	if p.StandardCost != nil {
		v := p.StandardCost.StringFixed(4)
		std = &v
	}
	_, err := s.db(tx).Exec(ctx, `
		UPDATE products SET name = $3, unit = $4, tracking_mode = $5, has_expiration = $6, valuation_method = $7,
			default_currency = $8, standard_cost = $9, updated_at = $10
		WHERE id = $1 AND tenant_code = $2
	`, p.ID, p.TenantCode, p.Name, p.Unit, p.TrackingMode, p.HasExpiration, p.ValuationMethod, p.DefaultCurrency, std, time.Now().UTC())
	return err
}

// DeleteProduct удаление.
func (s *Store) DeleteProduct(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM products WHERE id = $1 AND tenant_code = $2`, id, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ProductExistsSKU проверка sku.
func (s *Store) ProductExistsSKU(ctx context.Context, tx pgx.Tx, tenant, sku string) (bool, error) {
	var n int
	err := s.db(tx).QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE tenant_code = $1 AND sku = $2`, tenant, sku).Scan(&n)
	return n > 0, err
}

// --- Prices ---

// CreatePrice создаёт цену.
func (s *Store) CreatePrice(ctx context.Context, tx pgx.Tx, pr *models.ProductPrice) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO product_prices (id, tenant_code, product_id, price_type, currency, price, valid_from, valid_to)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, pr.ID, pr.TenantCode, pr.ProductID, pr.PriceType, pr.Currency, pr.Price.StringFixed(4), pr.ValidFrom, pr.ValidTo)
	return err
}

// GetPriceOnDate цена на дату.
func (s *Store) GetPriceOnDate(ctx context.Context, tx pgx.Tx, tenant string, productID uuid.UUID, priceType string, on time.Time) (*decimal.Decimal, error) {
	var p *string
	err := s.db(tx).QueryRow(ctx, `
		SELECT price::text FROM product_prices
		WHERE tenant_code = $1 AND product_id = $2 AND price_type = $3
		  AND valid_from <= $4::date AND (valid_to IS NULL OR valid_to >= $4::date)
		ORDER BY valid_from DESC LIMIT 1
	`, tenant, productID, priceType, on).Scan(&p)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	d, err := decimal.NewFromString(*p)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// OverlappingPrice проверка пересечения периодов.
func (s *Store) OverlappingPrice(ctx context.Context, tx pgx.Tx, tenant string, productID uuid.UUID, priceType string, from time.Time, to *time.Time, excludeID *uuid.UUID) (bool, error) {
	q := `
		SELECT COUNT(*) FROM product_prices
		WHERE tenant_code = $1 AND product_id = $2 AND price_type = $3
		  AND ($4::uuid IS NULL OR id <> $4::uuid)
		  AND NOT (COALESCE(valid_to, 'infinity'::date) < $5::date OR valid_from > COALESCE($6::date, 'infinity'::date))
	`
	var n int
	err := s.db(tx).QueryRow(ctx, q, tenant, productID, priceType, excludeID, from, to).Scan(&n)
	return n > 0, err
}

// ListPrices по товару.
func (s *Store) ListPrices(ctx context.Context, tx pgx.Tx, tenant string, productID uuid.UUID) ([]models.ProductPrice, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, product_id, price_type, currency, price::text, valid_from, valid_to, created_at
		FROM product_prices WHERE tenant_code = $1 AND product_id = $2 ORDER BY valid_from DESC
	`, tenant, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ProductPrice
	for rows.Next() {
		var pr models.ProductPrice
		var ps string
		if err := rows.Scan(&pr.ID, &pr.TenantCode, &pr.ProductID, &pr.PriceType, &pr.Currency, &ps, &pr.ValidFrom, &pr.ValidTo, &pr.CreatedAt); err != nil {
			return nil, err
		}
		pr.Price, _ = decimal.NewFromString(ps)
		out = append(out, pr)
	}
	return out, rows.Err()
}

// DeletePrice удаление цены.
func (s *Store) DeletePrice(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM product_prices WHERE id = $1 AND tenant_code = $2`, id, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
