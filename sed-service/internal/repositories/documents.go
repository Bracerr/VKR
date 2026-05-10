package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/sed-service/internal/models"
)

// CreateDocument создаёт документ.
func (s *Store) CreateDocument(ctx context.Context, tx pgx.Tx, d *models.Document) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO documents (id, tenant_code, type_id, number, title, status, author_sub, current_order_no, payload, warehouse_ref)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, d.ID, d.TenantCode, d.TypeID, d.Number, d.Title, d.Status, d.AuthorSub, d.CurrentOrderNo, d.Payload, d.WarehouseRef)
	return err
}

// GetDocument документ.
func (s *Store) GetDocument(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Document, error) {
	var d models.Document
	var wh []byte
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, type_id, number, title, status, author_sub, current_order_no, payload, warehouse_ref, created_at, updated_at
		FROM documents WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(
		&d.ID, &d.TenantCode, &d.TypeID, &d.Number, &d.Title, &d.Status, &d.AuthorSub, &d.CurrentOrderNo,
		&d.Payload, &wh, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(wh) > 0 {
		d.WarehouseRef = json.RawMessage(wh)
	}
	return &d, nil
}

// LockDocumentForUpdate блокирует строку документа.
func (s *Store) LockDocumentForUpdate(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.Document, error) {
	var d models.Document
	var wh []byte
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, type_id, number, title, status, author_sub, current_order_no, payload, warehouse_ref, created_at, updated_at
		FROM documents WHERE id = $1 AND tenant_code = $2 FOR UPDATE
	`, id, tenant).Scan(
		&d.ID, &d.TenantCode, &d.TypeID, &d.Number, &d.Title, &d.Status, &d.AuthorSub, &d.CurrentOrderNo,
		&d.Payload, &wh, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(wh) > 0 {
		d.WarehouseRef = json.RawMessage(wh)
	}
	return &d, nil
}

// UpdateDocument обновляет документ.
func (s *Store) UpdateDocument(ctx context.Context, tx pgx.Tx, d *models.Document) error {
	_, err := s.db(tx).Exec(ctx, `
		UPDATE documents SET title = $3, status = $4, current_order_no = $5, payload = $6, warehouse_ref = $7, updated_at = $8
		WHERE id = $1 AND tenant_code = $2
	`, d.ID, d.TenantCode, d.Title, d.Status, d.CurrentOrderNo, d.Payload, d.WarehouseRef, time.Now().UTC())
	return err
}

// ListDocuments список.
func (s *Store) ListDocuments(ctx context.Context, tx pgx.Tx, tenant string, status *string, authorSub *string) ([]models.Document, error) {
	q := `SELECT id, tenant_code, type_id, number, title, status, author_sub, current_order_no, payload, warehouse_ref, created_at, updated_at
		FROM documents WHERE tenant_code = $1`
	args := []interface{}{tenant}
	n := 2
	if status != nil {
		q += ` AND status = $` + strconv.Itoa(n)
		args = append(args, *status)
		n++
	}
	if authorSub != nil {
		q += ` AND author_sub = $` + strconv.Itoa(n)
		args = append(args, *authorSub)
	}
	q += ` ORDER BY created_at DESC LIMIT 500`
	rows, err := s.db(tx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func scanDocuments(rows pgx.Rows) ([]models.Document, error) {
	var out []models.Document
	for rows.Next() {
		var d models.Document
		var wh []byte
		if err := rows.Scan(&d.ID, &d.TenantCode, &d.TypeID, &d.Number, &d.Title, &d.Status, &d.AuthorSub, &d.CurrentOrderNo,
			&d.Payload, &wh, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		if len(wh) > 0 {
			d.WarehouseRef = json.RawMessage(wh)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListDocumentsPendingApproval документы, где пользователь может согласовать текущий этап.
func (s *Store) ListDocumentsPendingApproval(ctx context.Context, tx pgx.Tx, tenant, userSub string, roles []string) ([]models.Document, error) {
	// Пользователь подходит к шагу, если pending и (required_user_sub = sub ИЛИ required_role IN roles)
	rows, err := s.db(tx).Query(ctx, `
		SELECT DISTINCT d.id, d.tenant_code, d.type_id, d.number, d.title, d.status, d.author_sub, d.current_order_no, d.payload, d.warehouse_ref, d.created_at, d.updated_at
		FROM documents d
		JOIN document_approvals da ON da.document_id = d.id AND da.decision = 'PENDING'
		JOIN workflow_steps ws ON ws.id = da.step_id
		WHERE d.tenant_code = $1 AND d.status = 'IN_REVIEW'
		  AND ws.order_no = d.current_order_no
		  AND (
			(ws.required_user_sub IS NOT NULL AND ws.required_user_sub = $2)
			OR (ws.required_role IS NOT NULL AND ws.required_role = ANY($3::text[]))
		  )
		ORDER BY d.created_at DESC
		LIMIT 200
	`, tenant, userSub, roles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

// InsertDocumentApproval создаёт строку согласования.
func (s *Store) InsertDocumentApproval(ctx context.Context, tx pgx.Tx, id, docID, stepID uuid.UUID) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO document_approvals (id, document_id, step_id, decision) VALUES ($1,$2,$3,'PENDING')
	`, id, docID, stepID)
	return err
}

// ListDocumentApprovals список с join шагов.
func (s *Store) ListDocumentApprovals(ctx context.Context, tx pgx.Tx, docID uuid.UUID) ([]models.DocumentApproval, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT da.id, da.document_id, da.step_id, da.decision, da.decider_sub, da.comment, da.decided_at,
		       ws.order_no, ws.name, ws.required_role, ws.required_user_sub
		FROM document_approvals da
		JOIN workflow_steps ws ON ws.id = da.step_id
		WHERE da.document_id = $1
		ORDER BY ws.order_no, ws.id
	`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DocumentApproval
	for rows.Next() {
		var a models.DocumentApproval
		if err := rows.Scan(&a.ID, &a.DocumentID, &a.StepID, &a.Decision, &a.DeciderSub, &a.Comment, &a.DecidedAt,
			&a.OrderNo, &a.StepName, &a.RequiredRole, &a.RequiredUserSub); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// SetApprovalDecision фиксирует решение.
func (s *Store) SetApprovalDecision(ctx context.Context, tx pgx.Tx, approvalID uuid.UUID, decision, deciderSub, comment string) error {
	tag, err := s.db(tx).Exec(ctx, `
		UPDATE document_approvals SET decision = $2, decider_sub = $3, comment = $4, decided_at = now()
		WHERE id = $1 AND decision = 'PENDING'
	`, approvalID, decision, deciderSub, comment)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// DeleteDocumentApprovals удаляет все согласования документа.
func (s *Store) DeleteDocumentApprovals(ctx context.Context, tx pgx.Tx, docID uuid.UUID) error {
	_, err := s.db(tx).Exec(ctx, `DELETE FROM document_approvals WHERE document_id = $1`, docID)
	return err
}

// PendingApprovalsForOrder считает pending на этапе order_no.
func (s *Store) PendingApprovalsForOrder(ctx context.Context, tx pgx.Tx, docID uuid.UUID, orderNo int) (pending int, err error) {
	err = s.db(tx).QueryRow(ctx, `
		SELECT COUNT(*) FROM document_approvals da
		JOIN workflow_steps ws ON ws.id = da.step_id
		WHERE da.document_id = $1 AND da.decision = 'PENDING' AND ws.order_no = $2
	`, docID, orderNo).Scan(&pending)
	return pending, err
}

// MinPendingOrder минимальный order_no среди pending.
func (s *Store) MinPendingOrder(ctx context.Context, tx pgx.Tx, docID uuid.UUID) (*int, error) {
	var m sql.NullInt64
	err := s.db(tx).QueryRow(ctx, `
		SELECT MIN(ws.order_no) FROM document_approvals da
		JOIN workflow_steps ws ON ws.id = da.step_id
		WHERE da.document_id = $1 AND da.decision = 'PENDING'
	`, docID).Scan(&m)
	if err != nil {
		return nil, err
	}
	if !m.Valid {
		return nil, nil
	}
	v := int(m.Int64)
	return &v, nil
}

// FindPendingApprovalForUser находит одну pending-запись, которую может закрыть пользователь на текущем этапе.
func (s *Store) FindPendingApprovalForUser(ctx context.Context, tx pgx.Tx, docID uuid.UUID, orderNo int, userSub string, roles []string) (*uuid.UUID, error) {
	var aid uuid.UUID
	err := s.db(tx).QueryRow(ctx, `
		SELECT da.id FROM document_approvals da
		JOIN workflow_steps ws ON ws.id = da.step_id
		WHERE da.document_id = $1 AND da.decision = 'PENDING' AND ws.order_no = $2
		  AND (
			(ws.required_user_sub IS NOT NULL AND ws.required_user_sub = $3)
			OR (ws.required_role IS NOT NULL AND ws.required_role = ANY($4::text[]))
		  )
		ORDER BY ws.id LIMIT 1
	`, docID, orderNo, userSub, roles).Scan(&aid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &aid, nil
}

// InsertHistory аудит.
func (s *Store) InsertHistory(ctx context.Context, tx pgx.Tx, docID uuid.UUID, actorSub, action string, payload json.RawMessage) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO document_history (id, document_id, actor_sub, action, payload)
		VALUES ($1,$2,$3,$4,$5)
	`, uuid.New(), docID, actorSub, action, payload)
	return err
}
