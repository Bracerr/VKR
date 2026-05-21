package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEnvelope_and_TraceIngest(t *testing.T) {
	payload := map[string]string{"document_id": "d1"}
	env, err := NewEnvelope(TypeTraceDocumentPosted, "t1", "wh-doc-d1", payload)
	require.NoError(t, err)
	require.Equal(t, SpecVersion, env.SpecVersion)
	require.NotEmpty(t, env.EventID)

	legacy, err := env.ToTraceIngest()
	require.NoError(t, err)
	require.Equal(t, "DocumentPosted", legacy.EventType)
	require.Equal(t, "t1", legacy.TenantCode)
	require.NotNil(t, legacy.IdempotencyKey)
	require.Equal(t, "wh-doc-d1", *legacy.IdempotencyKey)

	b, err := env.Marshal()
	require.NoError(t, err)
	env2, err := Unmarshal(b)
	require.NoError(t, err)
	require.Equal(t, env.EventType, env2.EventType)

	var m map[string]any
	require.NoError(t, json.Unmarshal(legacy.Payload, &m))
	require.Equal(t, "d1", m["document_id"])
}

func TestSedPayloadRoundtrip(t *testing.T) {
	p := SedDocumentSignedPayload{DocumentID: "uuid", DocumentTypeCode: "PO_APPROVAL"}
	env, err := NewEnvelope(TypeSedDocumentSigned, "acme", "sed-sign-uuid", p)
	require.NoError(t, err)
	var decoded SedDocumentSignedPayload
	require.NoError(t, json.Unmarshal(env.Payload, &decoded))
	require.Equal(t, p.DocumentID, decoded.DocumentID)
}
