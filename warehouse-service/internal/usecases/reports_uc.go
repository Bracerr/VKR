package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// StockOnDate обёртка репозитория.
func (u *UC) StockOnDate(ctx context.Context, tenant string, at time.Time, whID, productID *uuid.UUID) ([]struct {
	WarehouseID, BinID, ProductID, BatchID uuid.UUID
	Qty                                     decimal.Decimal
	Value                                   decimal.Decimal
}, error) {
	return u.Store.StockOnDate(ctx, nil, tenant, at, whID, productID)
}

// Turnover обёртка.
func (u *UC) Turnover(ctx context.Context, tenant string, from, to time.Time, groupBy string) ([]struct {
	Key   string
	Qty   decimal.Decimal
	Value decimal.Decimal
}, error) {
	return u.Store.Turnover(ctx, nil, tenant, from, to, groupBy)
}

// ExpiringBatches скоро истекающие.
func (u *UC) ExpiringBatches(ctx context.Context, tenant string, before time.Time) ([]models.Batch, error) {
	return u.Store.ListExpiringBatches(ctx, nil, tenant, before)
}

// ABCRow строка ABC с классом.
type ABCRow struct {
	ProductID uuid.UUID `json:"product_id"`
	Metric    decimal.Decimal
	Class     string `json:"class"`
}

// ABCAnalysis классификация A/B/C по накопленной доле (80/95/100%).
func (u *UC) ABCAnalysis(ctx context.Context, tenant string, from, to time.Time, metric string) ([]ABCRow, error) {
	rows, err := u.Store.ABCAnalysis(ctx, nil, tenant, from, to, metric)
	if err != nil {
		return nil, err
	}
	var total decimal.Decimal
	for _, r := range rows {
		total = total.Add(r.Metric)
	}
	if total.IsZero() {
		out := make([]ABCRow, len(rows))
		for i, r := range rows {
			out[i] = ABCRow{ProductID: r.ProductID, Metric: r.Metric, Class: "C"}
		}
		return out, nil
	}
	var cum decimal.Decimal
	var out []ABCRow
	for _, r := range rows {
		cum = cum.Add(r.Metric)
		out = append(out, ABCRow{ProductID: r.ProductID, Metric: r.Metric, Class: abcClass(cum, total)})
	}
	return out, nil
}

func abcClass(cum, total decimal.Decimal) string {
	p80 := total.Mul(decimal.NewFromFloat(0.80))
	p95 := total.Mul(decimal.NewFromFloat(0.95))
	if cum.LessThanOrEqual(p80) {
		return "A"
	}
	if cum.LessThanOrEqual(p95) {
		return "B"
	}
	return "C"
}

// GetDocument складской документ.
func (u *UC) GetDocument(ctx context.Context, tenant string, id uuid.UUID) (*models.Document, error) {
	return u.Store.GetDocument(ctx, nil, tenant, id)
}

// ListMovements журнал.
func (u *UC) ListMovements(ctx context.Context, tenant string, from, to time.Time, whID, productID *uuid.UUID, movType *string, limit int) ([]models.StockMovement, error) {
	return u.Store.ListMovements(ctx, nil, tenant, from, to, whID, productID, movType, limit)
}

// ListBalances остатки.
func (u *UC) ListBalances(ctx context.Context, tenant string, whID, binID, productID, batchID *uuid.UUID, onlyPos bool, expBefore *time.Time) ([]models.StockBalance, error) {
	return u.Store.ListBalances(ctx, nil, tenant, whID, binID, productID, batchID, onlyPos, expBefore)
}

// PriceOnDate цена на дату.
func (u *UC) PriceOnDate(ctx context.Context, tenant string, productID uuid.UUID, priceType string, on time.Time) (*decimal.Decimal, error) {
	return u.Store.GetPriceOnDate(ctx, nil, tenant, productID, priceType, on)
}
