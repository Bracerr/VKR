package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/industrial-sed/production-service/internal/clients"
	"github.com/industrial-sed/production-service/internal/config"
	"github.com/industrial-sed/production-service/internal/repositories"
)

// Warehouse интеграция со складом (мок в тестах).
type Warehouse interface {
	CreateReservations(ctx context.Context, tenant, userName string, p *clients.WarehousePayload) ([]uuid.UUID, error)
	ReleaseReservation(ctx context.Context, tenant string, id uuid.UUID) error
	ConsumeReservation(ctx context.Context, tenant string, id uuid.UUID) error
	Receipt(ctx context.Context, tenant, userName string, p *clients.WarehousePayload) (uuid.UUID, error)
}

// SED интеграция с СЭД.
type SED interface {
	CreateDocument(ctx context.Context, bearer string, typeID uuid.UUID, title string, payload json.RawMessage) (*clients.SedDocument, error)
	SubmitDocument(ctx context.Context, bearer string, docID uuid.UUID) error
}

// App зависимости use cases.
type App struct {
	Store *repositories.Store
	WH    Warehouse
	SED   SED
	Trace *clients.Traceability
	Cfg   *config.Config
}
