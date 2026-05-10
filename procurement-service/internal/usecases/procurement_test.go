package usecases

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/procurement-service/internal/clients"
	"github.com/industrial-sed/procurement-service/internal/config"
	appmigrate "github.com/industrial-sed/procurement-service/internal/migrate"
	"github.com/industrial-sed/procurement-service/internal/repositories"
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
	lastTenant string
	lastReq    *clients.ReceiptRequest
	lastIdem   string
	docID      uuid.UUID
}

func (f *fakeWH) Receipt(_ context.Context, tenant string, req *clients.ReceiptRequest, idempotencyKey string) (uuid.UUID, error) {
	f.lastTenant, f.lastReq, f.lastIdem = tenant, req, idempotencyKey
	if f.docID == uuid.Nil {
		f.docID = uuid.New()
	}
	return f.docID, nil
}

func newAppForTests(t *testing.T) (*App, func()) {
	t.Helper()
	ctx := context.Background()
	dsn := "postgres://proc:proc@localhost:5436/procurement?sslmode=disable"
	pool, err := repositories.NewPool(ctx, dsn)
	if err != nil {
		t.Skip("нет локального Postgres :5436 для unit-тестов")
	}
	store := repositories.NewStore(pool)
	require.NoError(t, appmigrate.Up(dsn, "./migrations"))
	cleanup := func() { pool.Close() }
	return &App{Store: store, WH: &fakeWH{}, SED: &fakeSED{}, Cfg: &config.Config{}}, cleanup
}

func TestSubmitPR_SetsSubmittedAndSedID(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()

	tenant := "t1_" + uuid.New().String()[:8]
	pr, err := app.CreatePR(context.Background(), tenant, "u1", nil, nil)
	require.NoError(t, err)
	_, err = app.AddPRLine(context.Background(), tenant, "u1", pr.ID, 1, uuid.New(), "1", "pcs", nil, nil, nil)
	require.NoError(t, err)

	typeID := uuid.New()
	err = app.SubmitPR(context.Background(), tenant, "u1", "token", pr.ID, typeID, "PR approve")
	require.NoError(t, err)

	got, _, err := app.GetPRDetail(context.Background(), tenant, pr.ID)
	require.NoError(t, err)
	require.Equal(t, "SUBMITTED", got.Status)
	require.NotNil(t, got.SedDocumentID)

	sed := app.SED.(*fakeSED)
	require.Equal(t, typeID, sed.createdType)
	require.Len(t, sed.submitted, 1)
}

func TestSedCallback_ApprovesPRAndPO(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	tenant := "t2_" + uuid.New().String()[:8]

	// supplier
	sp, err := app.CreateSupplier(context.Background(), tenant, "u1", "S1", "Supplier", nil, nil, nil, true)
	require.NoError(t, err)

	// PR submitted
	pr, err := app.CreatePR(context.Background(), tenant, "u1", nil, nil)
	require.NoError(t, err)
	_, err = app.AddPRLine(context.Background(), tenant, "u1", pr.ID, 1, uuid.New(), "1", "pcs", nil, nil, nil)
	require.NoError(t, err)
	typeID := uuid.New()
	require.NoError(t, app.SubmitPR(context.Background(), tenant, "u1", "token", pr.ID, typeID, "PR approve"))
	pr2, _, _ := app.GetPRDetail(context.Background(), tenant, pr.ID)
	require.NotNil(t, pr2.SedDocumentID)

	require.NoError(t, app.HandleSedDocumentSigned(context.Background(), tenant, *pr2.SedDocumentID))
	pr3, _, _ := app.GetPRDetail(context.Background(), tenant, pr.ID)
	require.Equal(t, "APPROVED", pr3.Status)

	// PO submitted
	po, err := app.CreatePO(context.Background(), tenant, "u1", sp.ID, "RUB", nil, nil)
	require.NoError(t, err)
	_, err = app.AddPOLine(context.Background(), tenant, "u1", po.ID, 1, uuid.New(), "2", "0", "0", nil, nil)
	require.NoError(t, err)
	typeID2 := uuid.New()
	require.NoError(t, app.SubmitPO(context.Background(), tenant, "u1", "token", po.ID, typeID2, "PO approve"))
	po2, _, _ := app.GetPODetail(context.Background(), tenant, po.ID)
	require.NotNil(t, po2.SedDocumentID)

	require.NoError(t, app.HandleSedDocumentSigned(context.Background(), tenant, *po2.SedDocumentID))
	po3, _, _ := app.GetPODetail(context.Background(), tenant, po.ID)
	require.Equal(t, "APPROVED", po3.Status)
}

func TestReceivePO_UsesIdempotencyKeyAndSetsReceived(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	tenant := "t3_" + uuid.New().String()[:8]
	sp, err := app.CreateSupplier(context.Background(), tenant, "u1", "S1", "Supplier", nil, nil, nil, true)
	require.NoError(t, err)

	po, err := app.CreatePO(context.Background(), tenant, "u1", sp.ID, "RUB", nil, nil)
	require.NoError(t, err)
	_, err = app.AddPOLine(context.Background(), tenant, "u1", po.ID, 1, uuid.New(), "2", "0", "0", nil, nil)
	require.NoError(t, err)
	// имитируем что PO уже approved+released
	require.NoError(t, app.Store.UpdatePOStatusSed(context.Background(), nil, tenant, po.ID, "RELEASED", nil))

	whID := uuid.New()
	binID := uuid.New()
	docID, err := app.ReceivePO(context.Background(), tenant, "u1", po.ID, whID, binID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, docID)

	po2, _, _ := app.GetPODetail(context.Background(), tenant, po.ID)
	require.Equal(t, "RECEIVED", po2.Status)

	wh := app.WH.(*fakeWH)
	require.Contains(t, wh.lastIdem, po.ID.String())
	require.Equal(t, tenant, wh.lastTenant)
	require.Equal(t, whID.String(), wh.lastReq.WarehouseID)
	require.Equal(t, binID.String(), wh.lastReq.BinID)
}

func TestCrossTenant_CallbackDoesNotAffectOtherTenant(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	t1 := "t4_" + uuid.New().String()[:8]
	t2 := "t5_" + uuid.New().String()[:8]

	pr, err := app.CreatePR(context.Background(), t1, "u1", nil, nil)
	require.NoError(t, err)
	_, err = app.AddPRLine(context.Background(), t1, "u1", pr.ID, 1, uuid.New(), "1", "pcs", nil, nil, nil)
	require.NoError(t, err)
	typeID := uuid.New()
	require.NoError(t, app.SubmitPR(context.Background(), t1, "u1", "token", pr.ID, typeID, "PR approve"))

	pr2, _, _ := app.GetPRDetail(context.Background(), t1, pr.ID)
	require.NotNil(t, pr2.SedDocumentID)

	// callback with wrong tenant
	err = app.HandleSedDocumentSigned(context.Background(), t2, *pr2.SedDocumentID)
	require.Error(t, err)

	// still submitted in t1
	pr3, _, _ := app.GetPRDetail(context.Background(), t1, pr.ID)
	require.Equal(t, "SUBMITTED", pr3.Status)
}

func TestCancelPR_WrongState(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()
	tenant := "t6_" + uuid.New().String()[:8]
	pr, err := app.CreatePR(context.Background(), tenant, "u1", func() *time.Time { t := time.Now().UTC(); return &t }(), nil)
	require.NoError(t, err)
	require.NoError(t, app.Store.UpdatePRStatusSed(context.Background(), nil, tenant, pr.ID, "CANCELLED", nil))
	require.ErrorIs(t, app.CancelPR(context.Background(), tenant, "u1", pr.ID), ErrWrongState)
}

