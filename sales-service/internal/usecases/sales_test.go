package usecases

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/sales-service/internal/clients"
	appmigrate "github.com/industrial-sed/sales-service/internal/migrate"
	"github.com/industrial-sed/sales-service/internal/repositories"
)

type fakeSED struct {
	createdType uuid.UUID
	createdDoc  uuid.UUID
	submitted   []uuid.UUID
}

func (f *fakeSED) CreateDocument(_ context.Context, _ string, typeID uuid.UUID, _ string, _ json.RawMessage) (*clients.SedDocument, error) {
	f.createdType = typeID
	f.createdDoc = uuid.New()
	return &clients.SedDocument{ID: f.createdDoc, Status: "DRAFT"}, nil
}
func (f *fakeSED) SubmitDocument(_ context.Context, _ string, docID uuid.UUID) error {
	f.submitted = append(f.submitted, docID)
	return nil
}

type fakeWH struct {
	created []uuid.UUID
	issued  uuid.UUID
	released []uuid.UUID
}

func (f *fakeWH) CreateReservation(_ context.Context, _ string, _ *clients.ReservationRequest, _ string) (uuid.UUID, error) {
	id := uuid.New()
	f.created = append(f.created, id)
	return id, nil
}
func (f *fakeWH) ReleaseReservation(_ context.Context, _ string, id uuid.UUID) error {
	f.released = append(f.released, id)
	return nil
}
func (f *fakeWH) Issue(_ context.Context, _ string, _ *clients.IssueRequest, _ string) (uuid.UUID, error) {
	if f.issued == uuid.Nil {
		f.issued = uuid.New()
	}
	return f.issued, nil
}

func (f *fakeWH) IssueFromReservations(_ context.Context, _ string, _ []uuid.UUID, _ string) (uuid.UUID, error) {
	if f.issued == uuid.Nil {
		f.issued = uuid.New()
	}
	return f.issued, nil
}

func newAppForTests(t *testing.T) (*App, func()) {
	t.Helper()
	ctx := context.Background()
	dsn := "postgres://sales:sales@localhost:5437/sales?sslmode=disable"
	pool, err := repositories.NewPool(ctx, dsn)
	if err != nil {
		t.Skip("нет локального Postgres :5437 для unit-тестов")
	}
	require.NoError(t, appmigrate.Up(dsn, "./migrations"))
	store := repositories.NewStore(pool)
	return &App{Store: store, WH: &fakeWH{}, SED: &fakeSED{}}, func() { pool.Close() }
}

func TestSubmitSO_SetsSubmittedAndSedID(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	tenant := "t1_" + uuid.New().String()[:8]

	c, err := app.CreateCustomer(context.Background(), tenant, "u1", "C1", "Customer", nil, true)
	require.NoError(t, err)
	so, err := app.CreateSO(context.Background(), tenant, "u1", c.ID, nil, nil, nil)
	require.NoError(t, err)
	_, err = app.AddSOLine(context.Background(), tenant, "u1", so.ID, 1, uuid.New(), "2", "pcs", nil)
	require.NoError(t, err)

	typeID := uuid.New()
	require.NoError(t, app.SubmitSO(context.Background(), tenant, "u1", "token", so.ID, typeID, "SO approve"))

	got, _, err := app.GetSODetail(context.Background(), tenant, so.ID)
	require.NoError(t, err)
	require.Equal(t, "SUBMITTED", got.Status)
	require.NotNil(t, got.SedDocumentID)
}

func TestSedCallback_ApprovesSO(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	tenant := "t2_" + uuid.New().String()[:8]

	c, _ := app.CreateCustomer(context.Background(), tenant, "u1", "C1", "Customer", nil, true)
	so, _ := app.CreateSO(context.Background(), tenant, "u1", c.ID, nil, nil, nil)
	_, _ = app.AddSOLine(context.Background(), tenant, "u1", so.ID, 1, uuid.New(), "1", "pcs", nil)
	typeID := uuid.New()
	require.NoError(t, app.SubmitSO(context.Background(), tenant, "u1", "token", so.ID, typeID, "SO approve"))

	so2, _, _ := app.GetSODetail(context.Background(), tenant, so.ID)
	require.NotNil(t, so2.SedDocumentID)
	require.NoError(t, app.HandleSedDocumentSigned(context.Background(), tenant, *so2.SedDocumentID))
	so3, _, _ := app.GetSODetail(context.Background(), tenant, so.ID)
	require.Equal(t, "APPROVED", so3.Status)
}

func TestReserveAndShip(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	tenant := "t3_" + uuid.New().String()[:8]

	c, _ := app.CreateCustomer(context.Background(), tenant, "u1", "C1", "Customer", nil, true)
	whID := uuid.New()
	binID := uuid.New()
	so, _ := app.CreateSO(context.Background(), tenant, "u1", c.ID, &whID, &binID, nil)
	_, _ = app.AddSOLine(context.Background(), tenant, "u1", so.ID, 1, uuid.New(), "3", "pcs", nil)

	// approve+release
	require.NoError(t, app.Store.UpdateSOStatusSedAndReservations(context.Background(), nil, tenant, so.ID, "APPROVED", nil, []byte("[]")))
	require.NoError(t, app.ReleaseSO(context.Background(), tenant, "u1", so.ID))

	res, err := app.ReserveSO(context.Background(), tenant, "u1", so.ID)
	require.NoError(t, err)
	require.NotEmpty(t, res)

	docID, err := app.ShipSO(context.Background(), tenant, "u1", so.ID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, docID)

	got, _, _ := app.GetSODetail(context.Background(), tenant, so.ID)
	require.Equal(t, "SHIPPED", got.Status)
}

