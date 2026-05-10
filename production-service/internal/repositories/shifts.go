package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/production-service/internal/models"
)

// ListShiftTasks по тенанту и опциональной дате.
func (s *Store) ListShiftTasks(ctx context.Context, tx pgx.Tx, tenant string, date *time.Time) ([]models.ShiftTask, error) {
	q := `SELECT id, tenant_code, order_operation_id, shift_date, shift_no, assignee_sub,
		CASE WHEN qty_planned IS NULL THEN NULL ELSE qty_planned::text END, created_at
		FROM shift_tasks WHERE tenant_code=$1`
	args := []interface{}{tenant}
	if date != nil {
		q += ` AND shift_date=$2`
		args = append(args, *date)
	}
	q += ` ORDER BY shift_date DESC, shift_no`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ShiftTask
	for rows.Next() {
		var t models.ShiftTask
		if err := rows.Scan(&t.ID, &t.TenantCode, &t.OrderOperationID, &t.ShiftDate, &t.ShiftNo, &t.AssigneeSub, &t.QtyPlanned, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListShiftTasksForAssignee наряды пользователя.
func (s *Store) ListShiftTasksForAssignee(ctx context.Context, tx pgx.Tx, tenant, sub string) ([]models.ShiftTask, error) {
	q := `SELECT id, tenant_code, order_operation_id, shift_date, shift_no, assignee_sub,
		CASE WHEN qty_planned IS NULL THEN NULL ELSE qty_planned::text END, created_at
		FROM shift_tasks WHERE tenant_code=$1 AND assignee_sub=$2 ORDER BY shift_date DESC`
	rows, err := s.db(tx).Query(ctx, q, tenant, sub)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ShiftTask
	for rows.Next() {
		var t models.ShiftTask
		if err := rows.Scan(&t.ID, &t.TenantCode, &t.OrderOperationID, &t.ShiftDate, &t.ShiftNo, &t.AssigneeSub, &t.QtyPlanned, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CreateShiftTask создаёт.
func (s *Store) CreateShiftTask(ctx context.Context, tx pgx.Tx, t *models.ShiftTask) error {
	q := `INSERT INTO shift_tasks(id, tenant_code, order_operation_id, shift_date, shift_no, assignee_sub, qty_planned)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	var qp interface{}
	if t.QtyPlanned != nil && *t.QtyPlanned != "" {
		qp = *t.QtyPlanned
	}
	_, err := s.db(tx).Exec(ctx, q, t.ID, t.TenantCode, t.OrderOperationID, t.ShiftDate, t.ShiftNo, t.AssigneeSub, qp)
	return err
}

// DeleteShiftTask удаляет.
func (s *Store) DeleteShiftTask(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	q := `DELETE FROM shift_tasks WHERE tenant_code=$1 AND id=$2`
	_, err := s.db(tx).Exec(ctx, q, tenant, id)
	return err
}
