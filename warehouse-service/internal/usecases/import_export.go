package usecases

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// CreateImportJob создаёт задачу импорта товаров из CSV (строки: sku,name,unit,tracking_mode).
func (u *UC) CreateImportJob(ctx context.Context, tenant, user string, kind string, r io.Reader, maxRows int) (uuid.UUID, error) {
	if maxRows <= 0 {
		maxRows = 10000
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return uuid.Nil, err
	}
	cr := csv.NewReader(strings.NewReader(string(data)))
	cr.TrimLeadingSpace = true
	rows, err := cr.ReadAll()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: csv", ErrValidation)
	}
	if len(rows) < 2 {
		return uuid.Nil, fmt.Errorf("%w: пустой файл", ErrValidation)
	}
	tx, err := u.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	job := &models.ImportJob{
		ID:         uuid.New(),
		TenantCode: tenant,
		Kind:       kind,
		Status:     "QUEUED",
		Total:      len(rows) - 1,
		Processed:  0,
		CreatedBy:  user,
	}
	if job.Total > maxRows {
		return uuid.Nil, fmt.Errorf("%w: слишком много строк", ErrValidation)
	}
	if err := u.Store.CreateImportJob(ctx, tx, job); err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	go u.processProductImport(context.Background(), job.ID, tenant, rows[1:])
	return job.ID, nil
}

func (u *UC) processProductImport(ctx context.Context, jobID uuid.UUID, tenant string, rows [][]string) {
	_ = u.Store.UpdateImportJob(ctx, nil, jobID, "RUNNING", 0, nil)
	type errRow struct {
		Line int    `json:"line"`
		Msg  string `json:"msg"`
	}
	var errs []errRow
	done := 0
	for i, row := range rows {
		line := i + 2
		if len(row) < 4 {
			errs = append(errs, errRow{Line: line, Msg: "ожидаются колонки sku,name,unit,tracking_mode"})
			continue
		}
		sku, name, unit, tm := strings.TrimSpace(row[0]), strings.TrimSpace(row[1]), strings.TrimSpace(row[2]), strings.TrimSpace(row[3])
		if sku == "" || name == "" {
			errs = append(errs, errRow{Line: line, Msg: "sku и name обязательны"})
			continue
		}
		if unit == "" {
			unit = "pcs"
		}
		p := &models.Product{
			ID:              uuid.New(),
			TenantCode:      tenant,
			SKU:             sku,
			Name:            name,
			Unit:            unit,
			TrackingMode:    tm,
			HasExpiration:   false,
			ValuationMethod: models.ValAverage,
			DefaultCurrency: u.DefaultCurrency,
		}
		if p.TrackingMode == "" {
			p.TrackingMode = models.TrackingNone
		}
		if err := u.CreateProduct(ctx, p); err != nil {
			errs = append(errs, errRow{Line: line, Msg: err.Error()})
			continue
		}
		done++
	}
	status := "DONE"
	if len(errs) > 0 && done == 0 {
		status = "FAILED"
	}
	b, _ := json.Marshal(errs)
	_ = u.Store.UpdateImportJob(ctx, nil, jobID, status, done, b)
}

// GetImportJob статус задачи.
func (u *UC) GetImportJob(ctx context.Context, tenant string, id uuid.UUID) (*models.ImportJob, error) {
	return u.Store.GetImportJob(ctx, nil, tenant, id)
}

// ExportMovementsCSV потоковый экспорт движений в CSV.
func (u *UC) ExportMovementsCSV(ctx context.Context, w io.Writer, tenant string, from, to time.Time) error {
	movs, err := u.ListMovements(ctx, tenant, from, to, nil, nil, nil, 5000)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"id", "type", "warehouse_id", "bin_id", "product_id", "batch_id", "qty", "value", "posted_at"})
	for _, m := range movs {
		val := ""
		if m.Value != nil {
			val = m.Value.StringFixed(4)
		}
		_ = cw.Write([]string{
			m.ID.String(), m.MovementType, m.WarehouseID.String(), m.BinID.String(), m.ProductID.String(), m.BatchID.String(), m.Qty.StringFixed(3), val, m.PostedAt.Format(time.RFC3339),
		})
	}
	cw.Flush()
	return cw.Error()
}

// AverageCostOnDate средняя себестоимость по остатку на дату (value/qty по агрегату StockOnDate).
func (u *UC) AverageCostOnDate(ctx context.Context, tenant string, at time.Time, productID uuid.UUID) (*decimal.Decimal, error) {
	pid := productID
	rows, err := u.StockOnDate(ctx, tenant, at, nil, &pid)
	if err != nil {
		return nil, err
	}
	var q, v decimal.Decimal
	for _, r := range rows {
		q = q.Add(r.Qty)
		v = v.Add(r.Value)
	}
	if q.IsZero() {
		return nil, nil
	}
	x := v.Div(q)
	return &x, nil
}
