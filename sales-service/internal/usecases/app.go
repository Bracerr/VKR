package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/industrial-sed/sales-service/internal/clients"
	"github.com/industrial-sed/sales-service/internal/config"
	"github.com/industrial-sed/sales-service/internal/repositories"
)

type WarehouseIntegration interface {
	CreateReservation(ctx context.Context, tenant string, req *clients.ReservationRequest, idemKey string) (uuid.UUID, error)
	ReleaseReservation(ctx context.Context, tenant string, id uuid.UUID) error
	Issue(ctx context.Context, tenant string, req *clients.IssueRequest, idemKey string) (uuid.UUID, error)
	IssueFromReservations(ctx context.Context, tenant string, reservationIDs []uuid.UUID, idemKey string) (uuid.UUID, error)
}

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

