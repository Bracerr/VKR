package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/clients"
)

// WarehouseIntegration вызовы warehouse-service (для тестов — мок).
type WarehouseIntegration interface {
	CreateReservations(ctx context.Context, tenant, userName string, p *clients.WarehousePayload) ([]uuid.UUID, error)
	ConsumeReservations(ctx context.Context, tenant string, ids []uuid.UUID) error
	Receipt(ctx context.Context, tenant, userName string, p *clients.WarehousePayload) (uuid.UUID, error)
}

// RunWarehouseOnSign выполняет действие на складе по типу документа.
func RunWarehouseOnSign(ctx context.Context, tenant string, action string, payload json.RawMessage, wh WarehouseIntegration) (json.RawMessage, error) {
	switch action {
	case "NONE":
		return json.RawMessage(`{}`), nil
	case "RESERVE":
		if wh == nil {
			return nil, fmt.Errorf("warehouse client: %w", ErrValidation)
		}
		p, err := clients.ParseWarehousePayload(payload)
		if err != nil {
			return nil, err
		}
		ids, err := wh.CreateReservations(ctx, tenant, "", p)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrWarehouse, err)
		}
		out, _ := json.Marshal(map[string]any{"reservation_ids": ids})
		return out, nil
	case "CONSUME":
		if wh == nil {
			return nil, fmt.Errorf("warehouse client: %w", ErrValidation)
		}
		p, err := clients.ParseWarehousePayload(payload)
		if err != nil {
			return nil, err
		}
		var ids []uuid.UUID
		for _, s := range p.ReservationIDs {
			id, err := uuid.Parse(s)
			if err != nil {
				return nil, fmt.Errorf("reservation_ids: %w", ErrValidation)
			}
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return nil, fmt.Errorf("reservation_ids: %w", ErrValidation)
		}
		if err := wh.ConsumeReservations(ctx, tenant, ids); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrWarehouse, err)
		}
		out, _ := json.Marshal(map[string]any{"consumed_reservation_ids": p.ReservationIDs})
		return out, nil
	case "RECEIPT":
		if wh == nil {
			return nil, fmt.Errorf("warehouse client: %w", ErrValidation)
		}
		p, err := clients.ParseWarehousePayload(payload)
		if err != nil {
			return nil, err
		}
		docID, err := wh.Receipt(ctx, tenant, "", p)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrWarehouse, err)
		}
		out, _ := json.Marshal(map[string]any{"warehouse_document_id": docID})
		return out, nil
	default:
		return nil, fmt.Errorf("warehouse_action: %w", ErrValidation)
	}
}
