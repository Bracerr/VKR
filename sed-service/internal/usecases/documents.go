package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/sed-service/internal/models"
)

func histPayload(m map[string]any) json.RawMessage {
	b, _ := json.Marshal(m)
	return b
}

func warehouseRefSet(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	s := strings.TrimSpace(string(raw))
	return s != "" && s != "null"
}

// CreateDocument черновик.
func (a *App) CreateDocument(ctx context.Context, tenant, authorSub string, typeID uuid.UUID, title string, payload json.RawMessage) (*models.Document, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	dt, err := a.Store.GetDocumentType(ctx, tx, tenant, typeID)
	if err != nil {
		return nil, err
	}
	if dt == nil {
		return nil, ErrNotFound
	}
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	id := uuid.New()
	num := fmt.Sprintf("DOC-%s", strings.ReplaceAll(id.String(), "-", "")[:12])
	d := &models.Document{
		ID: id, TenantCode: tenant, TypeID: typeID, Number: num, Title: title,
		Status: "DRAFT", AuthorSub: authorSub, Payload: payload,
	}
	if err := a.Store.CreateDocument(ctx, tx, d); err != nil {
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, d.ID, authorSub, "CREATE", histPayload(map[string]any{"number": num})); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return d, nil
}

// GetDocument возвращает документ.
func (a *App) GetDocument(ctx context.Context, tenant string, id uuid.UUID) (*models.Document, error) {
	d, err := a.Store.GetDocument(ctx, nil, tenant, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

// ListDocuments список.
func (a *App) ListDocuments(ctx context.Context, tenant string, status *string, authorSub *string) ([]models.Document, error) {
	return a.Store.ListDocuments(ctx, nil, tenant, status, authorSub)
}

// UpdateDocument правка черновика автором.
func (a *App) UpdateDocument(ctx context.Context, tenant, actorSub string, id uuid.UUID, title *string, payload json.RawMessage) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.LockDocumentForUpdate(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	if d.Status != "DRAFT" {
		return ErrWrongState
	}
	if d.AuthorSub != actorSub {
		return ErrForbidden
	}
	if title != nil {
		d.Title = *title
	}
	if len(payload) > 0 {
		d.Payload = payload
	}
	if err := a.Store.UpdateDocument(ctx, tx, d); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, d.ID, actorSub, "UPDATE", histPayload(map[string]any{})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SubmitDocument отправка на согласование.
func (a *App) SubmitDocument(ctx context.Context, tenant, authorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.LockDocumentForUpdate(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	if d.Status != "DRAFT" {
		return ErrWrongState
	}
	if d.AuthorSub != authorSub {
		return ErrForbidden
	}
	dt, err := a.Store.GetDocumentType(ctx, tx, tenant, d.TypeID)
	if err != nil {
		return err
	}
	if dt == nil {
		return ErrNotFound
	}
	if dt.DefaultWorkflowID == nil {
		return fmt.Errorf("%w: у типа документа не задан маршрут", ErrValidation)
	}
	steps, err := a.Store.ListWorkflowSteps(ctx, tx, *dt.DefaultWorkflowID)
	if err != nil {
		return err
	}
	if len(steps) == 0 {
		return fmt.Errorf("%w: пустой маршрут", ErrValidation)
	}
	minOrder := steps[0].OrderNo
	for _, st := range steps[1:] {
		if st.OrderNo < minOrder {
			minOrder = st.OrderNo
		}
	}
	if err := a.Store.DeleteDocumentApprovals(ctx, tx, d.ID); err != nil {
		return err
	}
	for _, st := range steps {
		if err := a.Store.InsertDocumentApproval(ctx, tx, uuid.New(), d.ID, st.ID); err != nil {
			return err
		}
	}
	d.Status = "IN_REVIEW"
	d.CurrentOrderNo = &minOrder
	if err := a.Store.UpdateDocument(ctx, tx, d); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, d.ID, authorSub, "SUBMIT", histPayload(map[string]any{"current_order_no": minOrder})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ApproveDocument согласование текущего шага.
func (a *App) ApproveDocument(ctx context.Context, tenant, userSub string, roles []string, id uuid.UUID, comment string) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.LockDocumentForUpdate(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	if d.Status != "IN_REVIEW" || d.CurrentOrderNo == nil {
		return ErrWrongState
	}
	aid, err := a.Store.FindPendingApprovalForUser(ctx, tx, d.ID, *d.CurrentOrderNo, userSub, roles)
	if err != nil {
		return err
	}
	if aid == nil {
		return ErrForbidden
	}
	if err := a.Store.SetApprovalDecision(ctx, tx, *aid, "APPROVED", userSub, comment); err != nil {
		if err == pgx.ErrNoRows {
			return ErrConflict
		}
		return err
	}
	pending, err := a.Store.PendingApprovalsForOrder(ctx, tx, d.ID, *d.CurrentOrderNo)
	if err != nil {
		return err
	}
	if pending > 0 {
		if err := a.Store.InsertHistory(ctx, tx, d.ID, userSub, "APPROVE", histPayload(map[string]any{"parallel": true})); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	next, err := a.Store.MinPendingOrder(ctx, tx, d.ID)
	if err != nil {
		return err
	}
	if next == nil {
		d.Status = "APPROVED"
		d.CurrentOrderNo = nil
	} else {
		d.CurrentOrderNo = next
	}
	if err := a.Store.UpdateDocument(ctx, tx, d); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, d.ID, userSub, "APPROVE", histPayload(map[string]any{"next_order": d.CurrentOrderNo})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RejectDocument отклонение с возвратом в черновик.
func (a *App) RejectDocument(ctx context.Context, tenant, userSub string, roles []string, id uuid.UUID, comment string) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.LockDocumentForUpdate(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	if d.Status != "IN_REVIEW" || d.CurrentOrderNo == nil {
		return ErrWrongState
	}
	aid, err := a.Store.FindPendingApprovalForUser(ctx, tx, d.ID, *d.CurrentOrderNo, userSub, roles)
	if err != nil {
		return err
	}
	if aid == nil {
		return ErrForbidden
	}
	if err := a.Store.SetApprovalDecision(ctx, tx, *aid, "REJECTED", userSub, comment); err != nil {
		if err == pgx.ErrNoRows {
			return ErrConflict
		}
		return err
	}
	if err := a.Store.DeleteDocumentApprovals(ctx, tx, d.ID); err != nil {
		return err
	}
	d.Status = "DRAFT"
	d.CurrentOrderNo = nil
	if err := a.Store.UpdateDocument(ctx, tx, d); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, d.ID, userSub, "REJECT", histPayload(map[string]any{"comment": comment})); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// CancelDocument отмена.
func (a *App) CancelDocument(ctx context.Context, tenant, actorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.LockDocumentForUpdate(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	switch d.Status {
	case "DRAFT", "IN_REVIEW", "APPROVED":
	default:
		return ErrWrongState
	}
	if d.AuthorSub != actorSub {
		return ErrForbidden
	}
	_ = a.Store.DeleteDocumentApprovals(ctx, tx, d.ID)
	d.Status = "CANCELLED"
	d.CurrentOrderNo = nil
	if err := a.Store.UpdateDocument(ctx, tx, d); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, d.ID, actorSub, "CANCEL", nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SignDocument подпись и интеграция со складом.
func (a *App) SignDocument(ctx context.Context, tenant, authorSub string, id uuid.UUID) error {
	tx1, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	d, err := a.Store.LockDocumentForUpdate(ctx, tx1, tenant, id)
	if err != nil {
		_ = tx1.Rollback(ctx)
		return err
	}
	if d == nil {
		_ = tx1.Rollback(ctx)
		return ErrNotFound
	}
	if d.Status == "SIGNED" {
		return tx1.Commit(ctx)
	}
	if d.Status != "APPROVED" {
		_ = tx1.Rollback(ctx)
		return ErrWrongState
	}
	if d.AuthorSub != authorSub {
		_ = tx1.Rollback(ctx)
		return ErrForbidden
	}
	dt, err := a.Store.GetDocumentType(ctx, tx1, tenant, d.TypeID)
	if err != nil {
		_ = tx1.Rollback(ctx)
		return err
	}
	if dt == nil {
		_ = tx1.Rollback(ctx)
		return ErrNotFound
	}
	needWH := !warehouseRefSet(d.WarehouseRef)
	action := dt.WarehouseAction
	payload := append(json.RawMessage(nil), d.Payload...)
	existingRef := append(json.RawMessage(nil), d.WarehouseRef...)
	_ = tx1.Rollback(ctx)

	var whRef json.RawMessage
	if needWH {
		ref, err := RunWarehouseOnSign(ctx, tenant, action, payload, a.WH)
		if err != nil {
			return err
		}
		whRef = ref
	} else {
		whRef = existingRef
	}

	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	d2, err := a.Store.LockDocumentForUpdate(ctx, tx2, tenant, id)
	if err != nil {
		return err
	}
	if d2 == nil {
		return ErrNotFound
	}
	if d2.Status == "SIGNED" {
		return tx2.Commit(ctx)
	}
	if d2.Status != "APPROVED" {
		return ErrWrongState
	}
	if warehouseRefSet(d2.WarehouseRef) {
		whRef = append(json.RawMessage(nil), d2.WarehouseRef...)
	}
	d2.Status = "SIGNED"
	d2.WarehouseRef = whRef
	if err := a.Store.UpdateDocument(ctx, tx2, d2); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx2, d2.ID, authorSub, "SIGN", histPayload(map[string]any{"warehouse_action": action})); err != nil {
		return err
	}
	if err := tx2.Commit(ctx); err != nil {
		return err
	}
	if a.Prod != nil {
		bg, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = a.Prod.NotifyDocumentSigned(bg, tenant, d2.ID, dt.Code)
	}
	if a.Proc != nil {
		bg, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		_ = a.Proc.NotifyDocumentSigned(bg, tenant, d2.ID, dt.Code)
	}
	return nil
}

// ListDocumentApprovals история согласований.
func (a *App) ListDocumentApprovals(ctx context.Context, docID uuid.UUID) ([]models.DocumentApproval, error) {
	return a.Store.ListDocumentApprovals(ctx, nil, docID)
}

// ListDocumentHistory аудит.
func (a *App) ListDocumentHistory(ctx context.Context, docID uuid.UUID) ([]models.DocumentHistory, error) {
	return a.Store.ListDocumentHistory(ctx, nil, docID)
}

// ListTasks документы, ожидающие действия пользователя.
func (a *App) ListTasks(ctx context.Context, tenant, userSub string, roles []string) ([]models.Document, error) {
	return a.Store.ListDocumentsPendingApproval(ctx, nil, tenant, userSub, roles)
}
