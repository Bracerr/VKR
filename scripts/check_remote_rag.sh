#!/usr/bin/env bash
set -euo pipefail
BASE="${1:-http://85.236.191.21:52556}"
SECRET="${2:-e2e-test-secret}"

login() {
  curl -sS -m 15 -X POST "$BASE/api/v1/internal/test/login" \
    -H "X-Test-Secret: $SECRET" -H "Content-Type: application/json" \
    -d "{\"username\":\"$1\",\"password\":\"$2\"}"
}

echo "=== health ==="
curl -sS -m 8 "$BASE/health"; echo
curl -sS -m 8 "$BASE/ready"; echo

echo "=== ent-admin ==="
ENT_JSON=$(login "rag_ent_admin@ragcorp" "RagTest2026!")
ENT=$(echo "$ENT_JSON" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))")
test -n "$ENT" || { echo "ent-admin login failed: $ENT_JSON"; exit 1; }

mk() {
  local name="$1" role="$2"
  local r u p tok
  r=$(curl -sS -m 20 -X POST "$BASE/api/v1/users" \
    -H "Authorization: Bearer $ENT" -H "Content-Type: application/json" \
    -d "{\"username\":\"$name\",\"email\":\"${name}@t.local\",\"role\":\"$role\"}")
  u=$(echo "$r" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('username',''))" 2>/dev/null || echo "")
  p=$(echo "$r" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('temporary_password',''))" 2>/dev/null || echo "")
  if [ -z "$u" ] || [ -z "$p" ]; then
    echo "create user failed: $r" >&2
    return 1
  fi
  tok=$(login "$u" "$p" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || echo "")
  echo "$tok"
}

code() { curl -sS -m 12 -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $1" "$BASE/api/v1/documents/$2"; }
doc_count() { curl -sS -m 20 -H "Authorization: Bearer $1" "$BASE/api/v1/documents" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))"; }

TS=$(date +%s)
FIN=$(mk "rf_${TS}" "doc_read_finance")
SALES=$(mk "rs_${TS}" "doc_read_sales")
ADMIN=$(mk "ra_${TS}" "sed_admin")
NOACC=$(mk "rn_${TS}" "sed_viewer")

echo "=== counts: finance=$(doc_count "$FIN") sales=$(doc_count "$SALES") admin=$(doc_count "$ADMIN") ==="
echo "no_access list HTTP $(code "$NOACC" "00000000-0000-0000-0000-000000000001")" 
curl -sS -m 12 -o /dev/null -w "no_access list %{http_code}\n" -H "Authorization: Bearer $NOACC" "$BASE/api/v1/documents"

# resolve doc ids from admin list
DOC_JSON=$(curl -sS -m 25 -H "Authorization: Bearer $ADMIN" "$BASE/api/v1/documents")
PR_ID=$(echo "$DOC_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(next(x['id'] for x in d if 'RAG-PR' in x.get('title','')))")
SO_ID=$(echo "$DOC_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(next(x['id'] for x in d if 'RAG-SO' in x.get('title','')))")
SC_ID=$(echo "$DOC_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(next(x['id'] for x in d if 'RAG-SC' in x.get('title','')))")

echo "=== ACL by id ==="
echo "finance PR $PR_ID -> $(code "$FIN" "$PR_ID") (expect 200)"
echo "finance SO $SO_ID -> $(code "$FIN" "$SO_ID") (expect 403)"
echo "finance SC $SC_ID -> $(code "$FIN" "$SC_ID") (expect 200)"
echo "sales SO $SO_ID -> $(code "$SALES" "$SO_ID") (expect 200)"
echo "sales PR $PR_ID -> $(code "$SALES" "$PR_ID") (expect 403)"

RAG_PW="${RAG_FIXTURES_PASSWORD:-RagTest2026!}"
echo "=== rag_* fixed password ($RAG_PW) ==="
for u in rag_service rag_finance rag_admin; do
  code=$(curl -sS -m 10 -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/internal/test/login" \
    -H "X-Test-Secret: $SECRET" -H "Content-Type: application/json" \
    -d "{\"username\":\"${u}@ragcorp\",\"password\":\"$RAG_PW\"}")
  echo "$u@ragcorp -> $code"
done
SVC_TOKEN=$(login "rag_service@ragcorp" "$RAG_PW" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || echo "")
echo "rag_service doc count: $(doc_count "$SVC_TOKEN") (expect 50)"

echo "=== users count in ragcorp ==="
curl -sS -m 15 -H "Authorization: Bearer $ENT" "$BASE/api/v1/users" | python3 -c "import sys,json; print('users', len(json.load(sys.stdin)))"
