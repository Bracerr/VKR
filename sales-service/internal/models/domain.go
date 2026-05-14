package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID         uuid.UUID       `json:"id"`
	TenantCode string          `json:"tenant_code"`
	Code       string          `json:"code"`
	Name       string          `json:"name"`
	Contacts   json.RawMessage `json:"contacts"`
	Active     bool            `json:"active"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type SalesOrder struct {
	ID                uuid.UUID       `json:"id"`
	TenantCode         string          `json:"tenant_code"`
	Number             string          `json:"number"`
	Status             string          `json:"status"`
	CustomerID         uuid.UUID       `json:"customer_id"`
	CreatedBySub       string          `json:"created_by_sub"`
	ShipFromWarehouseID *uuid.UUID      `json:"ship_from_warehouse_id,omitempty"`
	ShipFromBinID       *uuid.UUID      `json:"ship_from_bin_id,omitempty"`
	Note               *string         `json:"note,omitempty"`
	SedDocumentID      *uuid.UUID      `json:"sed_document_id,omitempty"`
	Reservations       json.RawMessage `json:"reservations"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type SalesOrderLine struct {
	ID          uuid.UUID `json:"id"`
	SOID        uuid.UUID `json:"so_id"`
	LineNo      int       `json:"line_no"`
	ProductID   uuid.UUID `json:"product_id"`
	Qty         string    `json:"qty"`
	UOM         string    `json:"uom"`
	ReservedQty string    `json:"reserved_qty"`
	ShippedQty  string    `json:"shipped_qty"`
	Note        *string   `json:"note,omitempty"`
}

type Shipment struct {
	ID                uuid.UUID  `json:"id"`
	TenantCode        string     `json:"tenant_code"`
	SOID              uuid.UUID  `json:"so_id"`
	WarehouseDocumentID *uuid.UUID `json:"warehouse_document_id,omitempty"`
	Status            string     `json:"status"`
	PostedAt          time.Time  `json:"posted_at"`
}

