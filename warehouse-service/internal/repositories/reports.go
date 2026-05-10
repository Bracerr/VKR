package repositories

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// StockOnDateRow остаток на дату (агрегация движений).
func (s *Store) StockOnDate(ctx context.Context, tx pgx.Tx, tenant string, at time.Time, whID, productID *uuid.UUID) ([]struct {
	WarehouseID, BinID, ProductID, BatchID uuid.UUID
	Qty                                     decimal.Decimal
	Value                                   decimal.Decimal
}, error) {
	q := `
		SELECT warehouse_id, bin_id, product_id, batch_id,
		       SUM(qty)::text, COALESCE(SUM(value),0)::text
		FROM stock_movements
		WHERE tenant_code = $1 AND posted_at <= $2`
	args := []interface{}{tenant, at}
	n := 3
	if whID != nil {
		q += ` AND warehouse_id = $` + strconv.Itoa(n)
		args = append(args, *whID)
		n++
	}
	if productID != nil {
		q += ` AND product_id = $` + strconv.Itoa(n)
		args = append(args, *productID)
		n++
	}
	q += ` GROUP BY warehouse_id, bin_id, product_id, batch_id HAVING SUM(qty) <> 0`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		WarehouseID, BinID, ProductID, BatchID uuid.UUID
		Qty                                     decimal.Decimal
		Value                                   decimal.Decimal
	}
	for rows.Next() {
		var r struct {
			WarehouseID, BinID, ProductID, BatchID uuid.UUID
			Qty                                     decimal.Decimal
			Value                                   decimal.Decimal
		}
		var qs, vs string
		if err := rows.Scan(&r.WarehouseID, &r.BinID, &r.ProductID, &r.BatchID, &qs, &vs); err != nil {
			return nil, err
		}
		r.Qty, _ = decimal.NewFromString(qs)
		r.Value, _ = decimal.NewFromString(vs)
		out = append(out, r)
	}
	return out, rows.Err()
}

// TurnoverRow оборот.
func (s *Store) Turnover(ctx context.Context, tx pgx.Tx, tenant string, from, to time.Time, groupBy string) ([]struct {
	Key   string
	Qty   decimal.Decimal
	Value decimal.Decimal
}, error) {
	var sel string
	switch groupBy {
	case "warehouse":
		sel = "warehouse_id::text"
	case "batch":
		sel = "batch_id::text"
	case "bin":
		sel = "bin_id::text"
	default:
		sel = "product_id::text"
	}
	q := `SELECT ` + sel + ` AS k, SUM(ABS(qty))::text, COALESCE(SUM(ABS(COALESCE(value,0))),0)::text
		FROM stock_movements
		WHERE tenant_code = $1 AND posted_at >= $2 AND posted_at < $3 AND movement_type IN ('ISSUE','RECEIPT','TRANSFER_OUT','TRANSFER_IN','INVENTORY_ADJUST','RESERVE_CONSUMED')
		GROUP BY ` + sel + ` ORDER BY 2 DESC`
	rows, err := s.db(tx).Query(ctx, q, tenant, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Key   string
		Qty   decimal.Decimal
		Value decimal.Decimal
	}
	for rows.Next() {
		var r struct {
			Key   string
			Qty   decimal.Decimal
			Value decimal.Decimal
		}
		var qs, vs string
		if err := rows.Scan(&r.Key, &qs, &vs); err != nil {
			return nil, err
		}
		r.Qty, _ = decimal.NewFromString(qs)
		r.Value, _ = decimal.NewFromString(vs)
		out = append(out, r)
	}
	return out, rows.Err()
}

// ABCAnalysis оборот по товарам (классификацию A/B/C считает usecase по Парето).
func (s *Store) ABCAnalysis(ctx context.Context, tx pgx.Tx, tenant string, from, to time.Time, metric string) ([]struct {
	ProductID uuid.UUID
	Metric    decimal.Decimal
}, error) {
	valExpr := "SUM(ABS(qty))"
	if metric == "value" {
		valExpr = "SUM(ABS(COALESCE(value,0)))"
	}
	q := `SELECT product_id, ` + valExpr + `::text
		FROM stock_movements
		WHERE tenant_code = $1 AND posted_at >= $2 AND posted_at < $3
		GROUP BY product_id ORDER BY 2 DESC`
	rows, err := s.db(tx).Query(ctx, q, tenant, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		ProductID uuid.UUID
		Metric    decimal.Decimal
	}
	for rows.Next() {
		var r struct {
			ProductID uuid.UUID
			Metric    decimal.Decimal
		}
		var vs string
		if err := rows.Scan(&r.ProductID, &vs); err != nil {
			return nil, err
		}
		r.Metric, _ = decimal.NewFromString(vs)
		out = append(out, r)
	}
	return out, rows.Err()
}
