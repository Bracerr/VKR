package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/sed-service/internal/models"
)

// CreateWorkflow создаёт маршрут.
func (s *Store) CreateWorkflow(ctx context.Context, tx pgx.Tx, w *models.Workflow) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO workflows (id, tenant_code, code, name) VALUES ($1,$2,$3,$4)
	`, w.ID, w.TenantCode, w.Code, w.Name)
	return err
}

// GetWorkflow маршрут.
func (s *Store) GetWorkflow(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Workflow, error) {
	var w models.Workflow
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, code, name, created_at FROM workflows WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&w.ID, &w.TenantCode, &w.Code, &w.Name, &w.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

// ListWorkflows список.
func (s *Store) ListWorkflows(ctx context.Context, tx pgx.Tx, tenant string) ([]models.Workflow, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, code, name, created_at FROM workflows WHERE tenant_code = $1 ORDER BY code
	`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Workflow
	for rows.Next() {
		var w models.Workflow
		if err := rows.Scan(&w.ID, &w.TenantCode, &w.Code, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// UpdateWorkflow обновление.
func (s *Store) UpdateWorkflow(ctx context.Context, tx pgx.Tx, w *models.Workflow) error {
	tag, err := s.db(tx).Exec(ctx, `UPDATE workflows SET name = $3 WHERE id = $1 AND tenant_code = $2`, w.ID, w.TenantCode, w.Name)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteWorkflow удаление.
func (s *Store) DeleteWorkflow(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM workflows WHERE id = $1 AND tenant_code = $2`, id, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// AddWorkflowStep добавляет шаг.
func (s *Store) AddWorkflowStep(ctx context.Context, tx pgx.Tx, st *models.WorkflowStep) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO workflow_steps (id, workflow_id, order_no, parallel_group, name, required_role, required_user_sub)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, st.ID, st.WorkflowID, st.OrderNo, st.ParallelGroup, st.Name, st.RequiredRole, st.RequiredUserSub)
	return err
}

// ListWorkflowSteps шаги маршрута.
func (s *Store) ListWorkflowSteps(ctx context.Context, tx pgx.Tx, workflowID uuid.UUID) ([]models.WorkflowStep, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, workflow_id, order_no, parallel_group, name, required_role, required_user_sub, created_at
		FROM workflow_steps WHERE workflow_id = $1 ORDER BY order_no, id
	`, workflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.WorkflowStep
	for rows.Next() {
		var st models.WorkflowStep
		if err := rows.Scan(&st.ID, &st.WorkflowID, &st.OrderNo, &st.ParallelGroup, &st.Name, &st.RequiredRole, &st.RequiredUserSub, &st.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// DeleteWorkflowStep удаляет шаг.
func (s *Store) DeleteWorkflowStep(ctx context.Context, tx pgx.Tx, stepID uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM workflow_steps WHERE id = $1`, stepID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteWorkflowStepForTenant удаляет шаг, если маршрут принадлежит тенанту.
func (s *Store) DeleteWorkflowStepForTenant(ctx context.Context, tx pgx.Tx, tenant string, stepID uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `
		DELETE FROM workflow_steps ws USING workflows w
		WHERE ws.id = $1 AND ws.workflow_id = w.id AND w.tenant_code = $2
	`, stepID, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// --- document types ---

// CreateDocumentType создаёт тип.
func (s *Store) CreateDocumentType(ctx context.Context, tx pgx.Tx, dt *models.DocumentType) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO document_types (id, tenant_code, code, name, warehouse_action, default_workflow_id)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, dt.ID, dt.TenantCode, dt.Code, dt.Name, dt.WarehouseAction, dt.DefaultWorkflowID)
	return err
}

// GetDocumentType тип.
func (s *Store) GetDocumentType(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.DocumentType, error) {
	var dt models.DocumentType
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, code, name, warehouse_action, default_workflow_id, created_at
		FROM document_types WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&dt.ID, &dt.TenantCode, &dt.Code, &dt.Name, &dt.WarehouseAction, &dt.DefaultWorkflowID, &dt.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &dt, err
}

// ListDocumentTypes список.
func (s *Store) ListDocumentTypes(ctx context.Context, tx pgx.Tx, tenant string) ([]models.DocumentType, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, tenant_code, code, name, warehouse_action, default_workflow_id, created_at
		FROM document_types WHERE tenant_code = $1 ORDER BY code
	`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DocumentType
	for rows.Next() {
		var dt models.DocumentType
		if err := rows.Scan(&dt.ID, &dt.TenantCode, &dt.Code, &dt.Name, &dt.WarehouseAction, &dt.DefaultWorkflowID, &dt.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, dt)
	}
	return out, rows.Err()
}

// UpdateDocumentType обновление.
func (s *Store) UpdateDocumentType(ctx context.Context, tx pgx.Tx, dt *models.DocumentType) error {
	tag, err := s.db(tx).Exec(ctx, `
		UPDATE document_types SET name = $3, warehouse_action = $4, default_workflow_id = $5
		WHERE id = $1 AND tenant_code = $2
	`, dt.ID, dt.TenantCode, dt.Name, dt.WarehouseAction, dt.DefaultWorkflowID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteDocumentType удаление.
func (s *Store) DeleteDocumentType(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM document_types WHERE id = $1 AND tenant_code = $2`, id, tenant)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
