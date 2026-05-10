package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID         uuid.UUID       `json:"id"`
	TenantCode string          `json:"tenant_code"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	INN        *string         `json:"inn,omitempty"`
	KPP        *string         `json:"kpp,omitempty"`
	Contacts   json.RawMessage `json:"contacts"`
	Active     bool            `json:"active"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type PurchaseRequest struct {
	ID           uuid.UUID  `json:"id"`
	TenantCode   string     `json:"tenant_code"`
	Number       string     `json:"number"`
	Status       string     `json:"status"`
	CreatedBySub string     `json:"created_by_sub"`
	NeededBy     *time.Time `json:"needed_by,omitempty"`
	Note         *string    `json:"note,omitempty"`
	SedDocumentID *uuid.UUID `json:"sed_document_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PurchaseRequestLine struct {
	ID               uuid.UUID  `json:"id"`
	PRID             uuid.UUID  `json:"pr_id"`
	LineNo           int        `json:"line_no"`
	ProductID        uuid.UUID  `json:"product_id"`
	Qty              string     `json:"qty"`
	UOM              string     `json:"uom"`
	TargetWarehouseID *uuid.UUID `json:"target_warehouse_id,omitempty"`
	TargetBinID       *uuid.UUID `json:"target_bin_id,omitempty"`
	Note             *string    `json:"note,omitempty"`
}

type PurchaseOrder struct {
	ID           uuid.UUID  `json:"id"`
	TenantCode   string     `json:"tenant_code"`
	Number       string     `json:"number"`
	SupplierID   uuid.UUID  `json:"supplier_id"`
	Status       string     `json:"status"`
	CreatedBySub string     `json:"created_by_sub"`
	Currency     string     `json:"currency"`
	ExpectedAt   *time.Time `json:"expected_at,omitempty"`
	SedDocumentID *uuid.UUID `json:"sed_document_id,omitempty"`
	SourcePRID   *uuid.UUID `json:"source_pr_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PurchaseOrderLine struct {
	ID               uuid.UUID  `json:"id"`
	POID             uuid.UUID  `json:"po_id"`
	LineNo           int        `json:"line_no"`
	ProductID        uuid.UUID  `json:"product_id"`
	QtyOrdered       string     `json:"qty_ordered"`
	QtyReceived      string     `json:"qty_received"`
	Price            string     `json:"price"`
	VATRate          string     `json:"vat_rate"`
	TargetWarehouseID *uuid.UUID `json:"target_warehouse_id,omitempty"`
	TargetBinID       *uuid.UUID `json:"target_bin_id,omitempty"`
}

type Receipt struct {
	ID                uuid.UUID `json:"id"`
	TenantCode        string    `json:"tenant_code"`
	POID              uuid.UUID `json:"po_id"`
	WarehouseDocumentID uuid.UUID `json:"warehouse_document_id"`
	Status            string    `json:"status"`
	PostedAt          time.Time `json:"posted_at"`
}

