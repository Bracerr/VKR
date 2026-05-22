#!/usr/bin/env bash
# Синхронизация паролей rag_* в Keycloak (без удаления тенанта).
set -euo pipefail
BASE="${1:-http://85.236.191.21:52556}"
TENANT="${2:-ragcorp}"
PASSWORD="${RAG_FIXTURES_PASSWORD:-RagTest2026!}"
SECRET="${AUTH_SERVICE_SECRET:-e2e-service-secret}"

curl -sS -m 30 -X POST "$BASE/api/v1/internal/tenants/$TENANT/repair-passwords" \
  -H "X-Service-Secret: $SECRET" \
  -H "Content-Type: application/json" \
  -d "{\"password\":\"$PASSWORD\"}"
echo

for u in rag_service rag_finance rag_admin; do
  code=$(curl -sS -m 10 -o /tmp/t.json -w "%{http_code}" -X POST "$BASE/api/v1/internal/test/login" \
    -H "X-Test-Secret: ${AUTH_TEST_SECRET:-e2e-test-secret}" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"${u}@${TENANT}\",\"password\":\"$PASSWORD\"}")
  echo "$u@${TENANT} test login -> HTTP $code"
done
