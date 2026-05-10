package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// NilBatchID партия «без партии».
var NilBatchID = uuid.Nil

const (
	TrackingNone            = "NONE"
	TrackingBatch           = "BATCH"
	TrackingSerial          = "SERIAL"
	TrackingBatchAndSerial  = "BATCH_AND_SERIAL"
	ValFIFO                 = "FIFO"
	ValAverage              = "AVERAGE"
	ValStandard             = "STANDARD"
)

// Product товар.
type Product struct {
	ID               uuid.UUID
	TenantCode       string
	SKU              string
	Name             string
	Unit             string
	TrackingMode     string
	HasExpiration    bool
	ValuationMethod  string
	DefaultCurrency  string
	StandardCost     *decimal.Decimal
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Warehouse склад.
type Warehouse struct {
	ID         uuid.UUID
	TenantCode string
	Code       string
	Name       string
	CreatedAt  time.Time
}

// Bin ячейка.
type Bin struct {
	ID          uuid.UUID
	TenantCode  string
	WarehouseID uuid.UUID
	Code        string
	Name        string
	BinType     string
	ParentBinID *uuid.UUID
	CapacityQty *decimal.Decimal
	CreatedAt   time.Time
}

// Batch партия.
type Batch struct {
	ID             uuid.UUID
	TenantCode     string
	ProductID      uuid.UUID
	Series         string
	ManufacturedAt *time.Time
	ExpiresAt      *time.Time
	UnitCost       *decimal.Decimal
	Currency       *string
	CreatedAt      time.Time
}

// SerialNumber серийный номер.
type SerialNumber struct {
	ID             uuid.UUID
	TenantCode     string
	ProductID      uuid.UUID
	BatchID        *uuid.UUID
	SerialNo       string
	Status         string
	WarehouseID    *uuid.UUID
	BinID          *uuid.UUID
	LastMovementID *uuid.UUID
	UnitCost       *decimal.Decimal
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ProductPrice цена.
type ProductPrice struct {
	ID          uuid.UUID
	TenantCode  string
	ProductID   uuid.UUID
	PriceType   string
	Currency    string
	Price       decimal.Decimal
	ValidFrom   time.Time
	ValidTo     *time.Time
	CreatedAt   time.Time
}

// StockBalance остаток.
type StockBalance struct {
	WarehouseID uuid.UUID
	BinID       uuid.UUID
	ProductID   uuid.UUID
	BatchID     uuid.UUID
	Quantity    decimal.Decimal
	ReservedQty decimal.Decimal
	Value       decimal.Decimal
	UpdatedAt   time.Time
	ExpiresAt   *time.Time // из join batches для отчётов
}

// StockMovement движение.
type StockMovement struct {
	ID           uuid.UUID
	TenantCode   string
	MovementType string
	DocumentID   *uuid.UUID
	WarehouseID  uuid.UUID
	BinID        uuid.UUID
	ProductID    uuid.UUID
	BatchID      uuid.UUID
	SerialID     *uuid.UUID
	Qty          decimal.Decimal
	UnitCost     *decimal.Decimal
	Value        *decimal.Decimal
	Currency     *string
	PostedAt     time.Time
	PostedBy     string
}

// Document складской документ.
type Document struct {
	ID               uuid.UUID
	TenantCode       string
	DocType          string
	Number           string
	Status           string
	WarehouseFromID  *uuid.UUID
	WarehouseToID    *uuid.UUID
	PeriodAt         *time.Time
	CreatedBy        string
	CreatedAt        time.Time
}

// Reservation резерв.
type Reservation struct {
	ID          uuid.UUID
	TenantCode  string
	Status      string
	WarehouseID uuid.UUID
	BinID       *uuid.UUID
	ProductID   uuid.UUID
	BatchID     uuid.UUID
	SerialID    *uuid.UUID
	Qty         decimal.Decimal
	Reason      string
	DocRef      string
	ExpiresAt   *time.Time
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ImportJob задача импорта.
type ImportJob struct {
	ID          uuid.UUID
	TenantCode  string
	Kind        string
	Status      string
	Total       int
	Processed   int
	Errors      []byte // json
	CreatedBy   string
	CreatedAt   time.Time
	FinishedAt  *time.Time
}
