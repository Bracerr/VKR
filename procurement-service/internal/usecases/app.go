package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/industrial-sed/procurement-service/internal/clients"
	"github.com/industrial-sed/procurement-service/internal/config"
	"github.com/industrial-sed/procurement-service/internal/repositories"
)

// WarehouseIntegration calls.
type WarehouseIntegration interface {
	Receipt(ctx context.Context, tenant string, req *clients.ReceiptRequest, idempotencyKey string) (uuid.UUID, error)
}

// SedIntegration calls.
type SedIntegration interface {
	CreateDocument(ctx context.Context, bearer string, typeID uuid.UUID, title string, payload json.RawMessage) (*clients.SedDocument, error)
	SubmitDocument(ctx context.Context, bearer string, docID uuid.UUID) error
}

type App struct {
	Store *repositories.Store
	WH    WarehouseIntegration
	SED   SedIntegration
	Trace *clients.Traceability
	Cfg   *config.Config
}

