package outbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/industrial-sed/platform/events"
)

func TestBackoff(t *testing.T) {
	require.Equal(t, 2*time.Second, Backoff(1))
	require.True(t, Backoff(10) <= 60*time.Second)
}

func TestEnvelopeFromRow_roundtrip(t *testing.T) {
	env, err := events.NewEnvelope(events.TypeTraceDocumentPosted, "t1", "k1", map[string]string{"a": "b"})
	require.NoError(t, err)
	b, err := env.Marshal()
	require.NoError(t, err)
	row := Row{Payload: b}
	got, err := EnvelopeFromRow(row)
	require.NoError(t, err)
	require.Equal(t, env.EventType, got.EventType)
	require.Equal(t, env.TenantCode, got.TenantCode)
}
