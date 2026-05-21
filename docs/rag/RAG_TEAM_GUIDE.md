# RAG test pack — руководство для команды

Набор из **50 документов СЭД** (10 типов × 5), **10 тестовых пользователей** с разным ACL и JSON-manifest для индексации и проверки фильтрации по ролям.

## Включение fixtures

В `.env` корня репозитория:

```bash
RAG_FIXTURES_ENABLED=true
RAG_FIXTURES_TENANT=ragcorp
RAG_FIXTURES_PASSWORD=RagTest2026!
AUTH_BASE_URL=http://localhost:28080   # e2e; dev — свои URL
SED_BASE_URL=http://localhost:28091
WAREHOUSE_BASE_URL=http://localhost:28090
```

```bash
make test-up          # e2e-стенд
make seed-rag         # создать данные + manifest
make test-rag-fixtures  # прогон ACL-тестов
```

Отключение: `make seed-rag-teardown` (удаляет тенант `ragcorp` и generated JSON).

## Тенант и пользователи

| Username | Пароль | Роли | Видит чужих документов (ожидание) |
|----------|--------|------|-----------------------------------|
| `rag_admin` | см. `test_users.json` | `sed_admin`, `sed_author` | все 50 |
| `rag_finance` | * | `doc_read_finance` | 15 (закупки/договор) |
| `rag_procurement` | * | `doc_read_procurement` | 15 |
| `rag_sales` | * | `doc_read_sales` | 10 |
| `rag_production` | * | `doc_read_production` | 10 |
| `rag_warehouse` | * | `doc_read_warehouse` | 15 (складские типы) |
| `rag_author_proc` | * | `sed_author`, `doc_write_procurement` | свои + reader по типу |
| `rag_author_sales` | * | `sed_author`, `doc_write_sales` | аналогично |
| `rag_approver` | * | `sed_approver` | задачи `/tasks` + IN_REVIEW |
| `rag_no_access` | * | `sed_viewer` | 0 чужих |

Логин: `{username}@ragcorp` (например `rag_finance@ragcorp`). Актуальные пароли — в `docs/rag/generated/test_users.json` после `make seed-rag`.

## Файлы manifest

| Файл | Назначение |
|------|------------|
| `corpus_full.json` | **Полный корпус** для индексации: `search_text`, `payload`, ACL типа |
| `access_matrix.json` | Кто какие `document_id` может читать |
| `manifest_ids.json` | ID для автотестов и выборочных проверок |
| `test_users.json` | Учётки и пароли |
| `verification_report.json` | Результат сверки API при сиде |

## Индексация

1. Индексировать поле **`search_text`** (или `title` + `payload.description`) из `corpus_full.json`.
2. Для **per-user retrieval** фильтровать по `access_matrix.json` или запрашивать API под JWT нужного `rag_*`.
3. Уникальный ключ документа: `rag_id` в payload (`RAG-PR-001` … `RAG-WT-005`).

## Проверочные запросы (curl)

```bash
# Логин (подставьте пароль из test_users.json)
TOKEN=$(curl -s -X POST "$AUTH_BASE_URL/api/v1/internal/test/login" \
  -H "X-Test-Secret: e2e-test-secret" -H "Content-Type: application/json" \
  -d '{"username":"rag_finance@ragcorp","password":"..."}' | jq -r .access_token)

# Список (ожидается 15 документов)
curl -s -H "Authorization: Bearer $TOKEN" "$SED_BASE_URL/api/v1/documents" | jq 'length'

# Чужой SO — 403 (id из manifest_ids.json sample_so_doc_id)
curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" \
  "$SED_BASE_URL/api/v1/documents/<SO_UUID>"
```

## Негативные сценарии для RAG

- Запрос «договор поставщика RAG-SC» под `rag_finance` → находит; под `rag_sales` → не должен попасть в retrieval.
- `rag_no_access` — пустой список документов.

## Прогон автотестов перед передачей

```bash
make test-unit
make test-e2e-rag   # test-up + seed-rag + test-rag-fixtures
```

Ожидание: все тесты в `test_rag_fixtures.py` — **passed**.
