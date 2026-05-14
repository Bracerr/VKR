package usecases

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	appmigrate "github.com/industrial-sed/traceability-service/internal/migrate"
	"github.com/industrial-sed/traceability-service/internal/repositories"
)

func newAppForTests(t *testing.T) (*App, func()) {
	t.Helper()
	ctx := context.Background()
	dsn := "postgres://trace:trace@localhost:5438/trace?sslmode=disable"
	pool, err := repositories.NewPool(ctx, dsn)
	if err != nil {
		t.Skip("нет локального Postgres :5438 для unit-тестов")
	}
	require.NoError(t, appmigrate.Up(dsn, "./migrations"))
	store := repositories.NewStore(pool)
	return &App{Store: store}, func() { pool.Close() }
}

func TestIngest_DocumentPosted_CreatesNodesEdges(t *testing.T) {
	app, cleanup := newAppForTests(t)
	defer cleanup()

	tenant := "t_" + uuid.New().String()[:8]
	docID := uuid.New()
	batchID := uuid.New()
	serialID := uuid.New()
	now := time.Now().UTC()

	payload := DocumentPostedPayload{
		DocumentID: docID.String(),
		DocType:    "RECEIPT",
		Number:     "R-1",
		PostedAt:   now,
		Lines: []struct {
			ProductID    string  `json:"product_id"`
			BatchID      *string `json:"batch_id,omitempty"`
			BatchSeries  *string `json:"batch_series,omitempty"`
			SerialID     *string `json:"serial_id,omitempty"`
			SerialNo     *string `json:"serial_no,omitempty"`
			Qty          string  `json:"qty"`
		}{
			{
				ProductID: func() string { return uuid.New().String() }(),
				BatchID:   func() *string { s := batchID.String(); return &s }(),
				SerialID:  func() *string { s := serialID.String(); return &s }(),
				SerialNo:  func() *string { s := "SN-001"; return &s }(),
				Qty:       "1",
			},
		},
	}
	raw, _ := json.Marshal(payload)
	idem := "idem-1"
	require.NoError(t, app.Ingest(context.Background(), &IngestEvent{
		EventType:      EventDocumentPosted,
		TenantCode:     tenant,
		IdempotencyKey: &idem,
		Payload:        raw,
	}))

	docNode, err := app.Store.GetNodeID(context.Background(), nil, tenant, "WAREHOUSE_DOC", docID.String())
	require.NoError(t, err)
	require.NotNil(t, docNode)

	serNode, err := app.Store.GetNodeID(context.Background(), nil, tenant, "SERIAL", serialID.String())
	require.NoError(t, err)
	require.NotNil(t, serNode)
}

