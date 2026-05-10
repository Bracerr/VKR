package usecases

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/sed-service/internal/clients"
)

func TestRunWarehouseOnSign_NONE(t *testing.T) {
	raw, err := RunWarehouseOnSign(context.Background(), "t1", "NONE", json.RawMessage(`{}`), nil)
	require.NoError(t, err)
	require.JSONEq(t, `{}`, string(raw))
}

type fakeWh struct {
	ids      []uuid.UUID
	receipt  uuid.UUID
	err      error
	consumed int
}

func (f *fakeWh) CreateReservations(_ context.Context, _, _ string, _ *clients.WarehousePayload) ([]uuid.UUID, error) {
	return f.ids, f.err
}

func (f *fakeWh) ConsumeReservations(_ context.Context, _ string, ids []uuid.UUID) error {
	f.consumed += len(ids)
	return f.err
}

func (f *fakeWh) Receipt(_ context.Context, _, _ string, _ *clients.WarehousePayload) (uuid.UUID, error) {
	return f.receipt, f.err
}

func TestRunWarehouseOnSign_RESERVE_mock(t *testing.T) {
	id := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	f := &fakeWh{ids: []uuid.UUID{id}}
	whID := uuid.MustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	binID := uuid.MustParse("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
	pid := uuid.MustParse("6ba7b813-9dad-11d1-80b4-00c04fd430c8")
	raw, err := RunWarehouseOnSign(context.Background(), "t1", "RESERVE", json.RawMessage(`{
		"warehouse_id": "`+whID.String()+`",
		"default_bin_id": "`+binID.String()+`",
		"lines": [{"product_id": "`+pid.String()+`", "qty": "1", "reason": "t", "doc_ref": "d"}]
	}`), f)
	require.NoError(t, err)
	var out map[string][]uuid.UUID
	require.NoError(t, json.Unmarshal(raw, &out))
	require.Len(t, out["reservation_ids"], 1)
	require.Equal(t, id, out["reservation_ids"][0])
}

func TestRunWarehouseOnSign_CONSUME_mock(t *testing.T) {
	rid := uuid.MustParse("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
	f := &fakeWh{}
	raw, err := RunWarehouseOnSign(context.Background(), "t1", "CONSUME", json.RawMessage(`{
		"warehouse_id": "00000000-0000-0000-0000-000000000001",
		"reservation_ids": ["`+rid.String()+`"]
	}`), f)
	require.NoError(t, err)
	require.Equal(t, 1, f.consumed)
	require.Contains(t, string(raw), rid.String())
}

func TestRunWarehouseOnSign_nilClient(t *testing.T) {
	_, err := RunWarehouseOnSign(context.Background(), "t1", "RESERVE", json.RawMessage(`{}`), nil)
	require.ErrorIs(t, err, ErrValidation)
}

func TestRunWarehouseOnSign_badAction(t *testing.T) {
	_, err := RunWarehouseOnSign(context.Background(), "t1", "BOGUS", json.RawMessage(`{}`), &clients.Warehouse{})
	require.ErrorIs(t, err, ErrValidation)
}

func TestWarehouseRefSet(t *testing.T) {
	require.False(t, warehouseRefSet(nil))
	require.False(t, warehouseRefSet(json.RawMessage(`null`)))
	require.False(t, warehouseRefSet(json.RawMessage(`   `)))
	require.True(t, warehouseRefSet(json.RawMessage(`{}`)))
}
