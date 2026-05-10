package usecases

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/warehouse-service/internal/models"
	"github.com/industrial-sed/warehouse-service/internal/repositories"
)

// Интеграционные тесты: WAREHOUSE_TEST_DSN=postgres://...
func TestIntegrationReceiptIssueFEFO(t *testing.T) {
	dsn := os.Getenv("WAREHOUSE_TEST_DSN")
	if dsn == "" {
		t.Skip("WAREHOUSE_TEST_DSN not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	store := repositories.NewStore(pool)
	u := &UC{Store: store, DefaultCurrency: "RUB"}

	tenant := "t-" + uuid.NewString()[:8]
	user := "tester"

	w := &models.Warehouse{ID: uuid.New(), TenantCode: tenant, Code: "W1", Name: "Main"}
	require.NoError(t, u.CreateWarehouse(ctx, w))
	bin := &models.Bin{ID: uuid.New(), TenantCode: tenant, WarehouseID: w.ID, Code: "A1", Name: "Bin", BinType: "STORAGE"}
	require.NoError(t, u.CreateBin(ctx, bin))

	p := &models.Product{
		ID: uuid.New(), TenantCode: tenant, SKU: "SKU1", Name: "P1", Unit: "pcs",
		TrackingMode: models.TrackingBatch, ValuationMethod: models.ValAverage, DefaultCurrency: "RUB",
	}
	require.NoError(t, u.CreateProduct(ctx, p))

	e1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	e2 := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	s1 := "B-OLD"
	s2 := "B-NEW"
	c5, err := decimal.NewFromString("5")
	require.NoError(t, err)
	c7, err := decimal.NewFromString("7")
	require.NoError(t, err)

	lines := []ReceiptLineIn{
		{ProductID: p.ID, Qty: decimal.NewFromInt(10), Series: &s1, ExpiresAt: &e1, UnitCost: &c5},
		{ProductID: p.ID, Qty: decimal.NewFromInt(10), Series: &s2, ExpiresAt: &e2, UnitCost: &c7},
	}
	_, err = u.Receipt(ctx, tenant, user, w.ID, bin.ID, lines)
	require.NoError(t, err)

	docIssue, err := u.Issue(ctx, tenant, user, w.ID, bin.ID, []IssueIn{{ProductID: p.ID, Qty: decimal.NewFromInt(15)}})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, docIssue)

	_, err = u.Issue(ctx, tenant, user, w.ID, bin.ID, []IssueIn{{ProductID: p.ID, Qty: decimal.NewFromInt(10)}})
	require.ErrorIs(t, err, ErrInsufficient)
}
