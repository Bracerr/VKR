package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/sed-service/internal/models"
)

// --- Workflows ---

func (a *App) CreateWorkflow(ctx context.Context, tenant, code, name string) (*models.Workflow, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	w := &models.Workflow{ID: uuid.New(), TenantCode: tenant, Code: code, Name: name}
	if err := a.Store.CreateWorkflow(ctx, tx, w); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (a *App) ListWorkflows(ctx context.Context, tenant string) ([]models.Workflow, error) {
	return a.Store.ListWorkflows(ctx, nil, tenant)
}

func (a *App) UpdateWorkflow(ctx context.Context, tenant string, id uuid.UUID, name string) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	w, err := a.Store.GetWorkflow(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if w == nil {
		return ErrNotFound
	}
	w.Name = name
	if err := a.Store.UpdateWorkflow(ctx, tx, w); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

func (a *App) DeleteWorkflow(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteWorkflow(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

// AddWorkflowStep добавляет шаг (роль или sub обязательны снаружи).
func (a *App) AddWorkflowStep(ctx context.Context, tenant string, workflowID uuid.UUID, orderNo int, parallelGroup *int, name string, role, userSub *string) (*models.WorkflowStep, error) {
	if (role == nil || *role == "") && (userSub == nil || *userSub == "") {
		return nil, fmt.Errorf("%w: required_role или required_user_sub", ErrValidation)
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	wf, err := a.Store.GetWorkflow(ctx, tx, tenant, workflowID)
	if err != nil {
		return nil, err
	}
	if wf == nil {
		return nil, ErrNotFound
	}
	st := &models.WorkflowStep{
		ID: uuid.New(), WorkflowID: workflowID, OrderNo: orderNo, ParallelGroup: parallelGroup,
		Name: name, RequiredRole: role, RequiredUserSub: userSub,
	}
	if err := a.Store.AddWorkflowStep(ctx, tx, st); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return st, nil
}

func (a *App) ListWorkflowSteps(ctx context.Context, tenant string, workflowID uuid.UUID) ([]models.WorkflowStep, error) {
	wf, err := a.Store.GetWorkflow(ctx, nil, tenant, workflowID)
	if err != nil {
		return nil, err
	}
	if wf == nil {
		return nil, ErrNotFound
	}
	return a.Store.ListWorkflowSteps(ctx, nil, workflowID)
}

func (a *App) DeleteWorkflowStep(ctx context.Context, tenant string, stepID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteWorkflowStepForTenant(ctx, tx, tenant, stepID); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

// --- Document types ---

func (a *App) CreateDocumentType(ctx context.Context, tenant, code, name, whAction string, workflowID *uuid.UUID) (*models.DocumentType, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if workflowID != nil {
		wf, err := a.Store.GetWorkflow(ctx, tx, tenant, *workflowID)
		if err != nil {
			return nil, err
		}
		if wf == nil {
			return nil, ErrNotFound
		}
	}
	dt := &models.DocumentType{
		ID: uuid.New(), TenantCode: tenant, Code: code, Name: name,
		WarehouseAction: whAction, DefaultWorkflowID: workflowID,
	}
	if dt.WarehouseAction == "" {
		dt.WarehouseAction = "NONE"
	}
	if err := a.Store.CreateDocumentType(ctx, tx, dt); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return dt, nil
}

func (a *App) ListDocumentTypes(ctx context.Context, tenant string) ([]models.DocumentType, error) {
	return a.Store.ListDocumentTypes(ctx, nil, tenant)
}

func (a *App) GetDocumentType(ctx context.Context, tenant string, id uuid.UUID) (*models.DocumentType, error) {
	dt, err := a.Store.GetDocumentType(ctx, nil, tenant, id)
	if err != nil {
		return nil, err
	}
	if dt == nil {
		return nil, ErrNotFound
	}
	return dt, nil
}

func (a *App) UpdateDocumentType(ctx context.Context, tenant string, id uuid.UUID, name, whAction string, workflowID *uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if workflowID != nil {
		wf, err := a.Store.GetWorkflow(ctx, tx, tenant, *workflowID)
		if err != nil {
			return err
		}
		if wf == nil {
			return ErrNotFound
		}
	}
	dt, err := a.Store.GetDocumentType(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if dt == nil {
		return ErrNotFound
	}
	dt.Name = name
	dt.WarehouseAction = whAction
	dt.DefaultWorkflowID = workflowID
	if err := a.Store.UpdateDocumentType(ctx, tx, dt); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}

func (a *App) DeleteDocumentType(ctx context.Context, tenant string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteDocumentType(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return tx.Commit(ctx)
}
