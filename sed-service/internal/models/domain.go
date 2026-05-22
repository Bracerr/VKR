package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DocumentType тип документа.
type DocumentType struct {
	ID                 uuid.UUID  `json:"id"`
	TenantCode         string     `json:"tenant_code"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	WarehouseAction    string     `json:"warehouse_action"`
	DefaultWorkflowID  *uuid.UUID `json:"default_workflow_id,omitempty"`
	ReaderRoles        []string   `json:"reader_roles"`
	WriterRoles        []string   `json:"writer_roles"`
	CreatedAt          time.Time  `json:"created_at"`
}

// Workflow маршрут согласования.
type Workflow struct {
	ID         uuid.UUID `json:"id"`
	TenantCode string    `json:"tenant_code"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

// WorkflowStep шаг.
type WorkflowStep struct {
	ID               uuid.UUID `json:"id"`
	WorkflowID       uuid.UUID `json:"workflow_id"`
	OrderNo          int       `json:"order_no"`
	ParallelGroup    *int      `json:"parallel_group,omitempty"`
	Name             string    `json:"name"`
	RequiredRole     *string   `json:"required_role,omitempty"`
	RequiredUserSub  *string   `json:"required_user_sub,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Document документ СЭД.
type Document struct {
	ID               uuid.UUID       `json:"id"`
	TenantCode       string          `json:"tenant_code"`
	TypeID           uuid.UUID       `json:"type_id"`
	Number           string          `json:"number"`
	Title            string          `json:"title"`
	Status           string          `json:"status"`
	AuthorSub        string          `json:"author_sub"`
	CurrentOrderNo   *int            `json:"current_order_no,omitempty"`
	Payload          json.RawMessage `json:"payload"`
	RagContent       json.RawMessage `json:"-"`
	WarehouseRef     json.RawMessage `json:"warehouse_ref,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// DocumentPublic — карточка документа для пользовательского API (без rag_content).
type DocumentPublic struct {
	ID             uuid.UUID       `json:"id"`
	TenantCode     string          `json:"tenant_code"`
	TypeID         uuid.UUID       `json:"type_id"`
	Number         string          `json:"number"`
	Title          string          `json:"title"`
	Status         string          `json:"status"`
	AuthorSub      string          `json:"author_sub"`
	CurrentOrderNo *int            `json:"current_order_no,omitempty"`
	Payload        json.RawMessage `json:"payload"`
	WarehouseRef   json.RawMessage `json:"warehouse_ref,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ToDocumentPublic убирает поля индексации RAG из ответа API.
func ToDocumentPublic(d *Document) DocumentPublic {
	if d == nil {
		return DocumentPublic{}
	}
	return DocumentPublic{
		ID: d.ID, TenantCode: d.TenantCode, TypeID: d.TypeID, Number: d.Number,
		Title: d.Title, Status: d.Status, AuthorSub: d.AuthorSub, CurrentOrderNo: d.CurrentOrderNo,
		Payload: d.Payload, WarehouseRef: d.WarehouseRef, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

// DocumentApproval строка согласования.
type DocumentApproval struct {
	ID          uuid.UUID  `json:"id"`
	DocumentID  uuid.UUID  `json:"document_id"`
	StepID      uuid.UUID  `json:"step_id"`
	Decision    string     `json:"decision"`
	DeciderSub  *string    `json:"decider_sub,omitempty"`
	Comment     *string    `json:"comment,omitempty"`
	DecidedAt   *time.Time `json:"decided_at,omitempty"`
	OrderNo     int        `json:"order_no"` // join from step
	StepName    string     `json:"step_name"`
	RequiredRole     *string `json:"required_role,omitempty"`
	RequiredUserSub  *string `json:"required_user_sub,omitempty"`
}

// DocumentHistory запись аудита.
type DocumentHistory struct {
	ID         uuid.UUID       `json:"id"`
	DocumentID uuid.UUID       `json:"document_id"`
	ActorSub   string          `json:"actor_sub"`
	Action     string          `json:"action"`
	Payload    json.RawMessage `json:"payload,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// DocumentFile вложение.
type DocumentFile struct {
	ID           uuid.UUID `json:"id"`
	DocumentID   uuid.UUID `json:"document_id"`
	ObjectKey    string    `json:"object_key"`
	OriginalName string    `json:"original_name"`
	ContentType  *string   `json:"content_type,omitempty"`
	SizeBytes    int64     `json:"size_bytes"`
	UploadedBy   string    `json:"uploaded_by"`
	UploadedAt   time.Time `json:"uploaded_at"`
}
