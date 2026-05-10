package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Workcenter рабочий центр.
type Workcenter struct {
	ID                      uuid.UUID `json:"id"`
	TenantCode              string    `json:"tenant_code"`
	Code                    string    `json:"code"`
	Name                    string    `json:"name"`
	Active                  bool      `json:"active"`
	CapacityMinutesPerShift *int      `json:"capacity_minutes_per_shift"`
	CreatedAt               time.Time `json:"created_at"`
}

// ScrapReason причина брака.
type ScrapReason struct {
	ID         uuid.UUID `json:"id"`
	TenantCode string    `json:"tenant_code"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
}

// BOM спецификация.
type BOM struct {
	ID            uuid.UUID  `json:"id"`
	TenantCode    string     `json:"tenant_code"`
	ProductID     uuid.UUID  `json:"product_id"`
	Version       int        `json:"version"`
	Status        string     `json:"status"`
	SedDocumentID *uuid.UUID `json:"sed_document_id"`
	ValidFrom     *time.Time `json:"valid_from"`
	ValidTo       *time.Time `json:"valid_to"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// BOMLine строка BOM.
type BOMLine struct {
	ID                 uuid.UUID  `json:"id"`
	BomID              uuid.UUID  `json:"bom_id"`
	LineNo             int        `json:"line_no"`
	ComponentProductID uuid.UUID  `json:"component_product_id"`
	QtyPer             string     `json:"qty_per"`
	ScrapPct           string     `json:"scrap_pct"`
	OpNo               int        `json:"op_no"`
	AltGroup           *string    `json:"alt_group"`
}

// Routing техкарта.
type Routing struct {
	ID            uuid.UUID  `json:"id"`
	TenantCode    string     `json:"tenant_code"`
	ProductID     uuid.UUID  `json:"product_id"`
	Version       int        `json:"version"`
	Status        string     `json:"status"`
	SedDocumentID *uuid.UUID `json:"sed_document_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// RoutingOperation операция маршрута.
type RoutingOperation struct {
	ID             uuid.UUID `json:"id"`
	RoutingID      uuid.UUID `json:"routing_id"`
	OpNo           int       `json:"op_no"`
	WorkcenterID   uuid.UUID `json:"workcenter_id"`
	Name           string    `json:"name"`
	TimePerUnitMin *string   `json:"time_per_unit_min"`
	SetupTimeMin   *string   `json:"setup_time_min"`
	QCRequired     bool      `json:"qc_required"`
}

// ProductionOrder заказ.
type ProductionOrder struct {
	ID                     uuid.UUID       `json:"id"`
	TenantCode             string          `json:"tenant_code"`
	Code                   string          `json:"code"`
	ProductID              uuid.UUID       `json:"product_id"`
	QtyPlanned             string          `json:"qty_planned"`
	QtyDone                string          `json:"qty_done"`
	QtyScrap               string          `json:"qty_scrap"`
	Status                 string          `json:"status"`
	BomID                  uuid.UUID       `json:"bom_id"`
	RoutingID              uuid.UUID       `json:"routing_id"`
	WarehouseID            uuid.UUID       `json:"warehouse_id"`
	DefaultBinID           uuid.UUID       `json:"default_bin_id"`
	Reservations           json.RawMessage `json:"reservations"`
	WarehouseReceiptDocID  *uuid.UUID      `json:"warehouse_receipt_doc_id"`
	StartPlan              *time.Time      `json:"start_plan"`
	FinishPlan             *time.Time      `json:"finish_plan"`
	StartFact              *time.Time      `json:"start_fact"`
	FinishFact             *time.Time      `json:"finish_fact"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

// OrderOperationSnapshot операция в заказе.
type OrderOperation struct {
	ID            uuid.UUID  `json:"id"`
	OrderID       uuid.UUID  `json:"order_id"`
	OpNo          int        `json:"op_no"`
	WorkcenterID  uuid.UUID  `json:"workcenter_id"`
	Name          string     `json:"name"`
	QtyPlanned    string     `json:"qty_planned"`
	QtyGood       string     `json:"qty_good"`
	QtyScrap      string     `json:"qty_scrap"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
}

// ShiftTask сменное задание.
type ShiftTask struct {
	ID                uuid.UUID  `json:"id"`
	TenantCode        string     `json:"tenant_code"`
	OrderOperationID  uuid.UUID  `json:"order_operation_id"`
	ShiftDate         time.Time  `json:"shift_date"`
	ShiftNo           int        `json:"shift_no"`
	AssigneeSub       *string    `json:"assignee_sub"`
	QtyPlanned        *string    `json:"qty_planned"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ProductionReport факт.
type ProductionReport struct {
	ID                 uuid.UUID `json:"id"`
	TenantCode         string    `json:"tenant_code"`
	OrderOperationID   uuid.UUID `json:"order_operation_id"`
	ReportedBySub      string    `json:"reported_by_sub"`
	QtyGood            string    `json:"qty_good"`
	QtyScrap           string    `json:"qty_scrap"`
	ScrapReasonCode    *string   `json:"scrap_reason_code"`
	Note               *string   `json:"note"`
	CreatedAt          time.Time `json:"created_at"`
}

// ReservationLine связь линии BOM с резервом на складе.
type ReservationLine struct {
	BomLineID     uuid.UUID `json:"bom_line_id"`
	ReservationID uuid.UUID `json:"reservation_id"`
}
