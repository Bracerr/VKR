# RAG — руководство для команды

Корпус из **50 документов СЭД** (10 типов × 5), **10 учётных записей** с разным ACL и JSON-manifest для индексации и проверки фильтрации по ролям.

**Тенант:** `ragcorp`  
**Формат логина:** `{username}@ragcorp`

**Базовый URL стенда (один порт для всего API и Keycloak):**

```text
http://<хост>:5656
```

Все запросы ниже идут на этот адрес (префиксы `/api/v1/...`, `/realms/...`).

Realm: `industrial-sed`. Client для machine-login: `auth-service` (уточните `client_secret` у администратора).

---

## Учётные записи

Пароли и `user_id` — в файле `docs/rag/generated/test_users.json` (поставляется вместе с manifest).

| Username | Логин | Пароль | Роли | `GET /api/v1/documents` |
|----------|-------|--------|------|-------------------------|
| `rag_admin` | `rag_admin@ragcorp` | `jHyZyS5nQZoFq1ja` | `sed_admin`, `sed_author` | **200**, **50** записей |
| `rag_finance` | `rag_finance@ragcorp` | `CM3L5KExDQQ9oxqQ` | `doc_read_finance` | **200**, **15** (PR, PO, SC) |
| `rag_procurement` | `rag_procurement@ragcorp` | `YJfMYLStdw%Qg28#` | `doc_read_procurement` | **200**, **15** (PR, PO, SC) |
| `rag_sales` | `rag_sales@ragcorp` | `kDBcCHSx7o3ueAob` | `doc_read_sales` | **200**, **10** (SO, SH) |
| `rag_production` | `rag_production@ragcorp` | `fI9Eg$QPw75DN7EE` | `doc_read_production` | **200**, **10** (BOM, RT) |
| `rag_warehouse` | `rag_warehouse@ragcorp` | `V6m8QzyLk1OentGE` | `doc_read_warehouse` | **200**, **15** (WR, WC, WT) |
| `rag_author_proc` | `rag_author_proc@ragcorp` | `g2XlmkXkvUxYRGgL` | `sed_author`, `doc_write_procurement` | см. сценарии ниже |
| `rag_author_sales` | `rag_author_sales@ragcorp` | `eBDVl1LMcU644ggL` | `sed_author`, `doc_write_sales` | см. сценарии ниже |
| `rag_approver` | `rag_approver@ragcorp` | `!k3lboMwVd3qMnGR` | `sed_approver` | **200**, **0**; задачи — `GET /tasks` |
| `rag_no_access` | `rag_no_access@ragcorp` | `#NZC!B7Iu$0o%8dv` | `sed_viewer` | **403** |

### Какие типы документов видит роль

| Код типа | Префикс `rag_id` | Кто читает (`doc_read_*`) |
|----------|------------------|---------------------------|
| `PURCHASE_REQUEST_APPROVAL` | `RAG-PR-xxx` | procurement, finance |
| `PURCHASE_ORDER_APPROVAL` | `RAG-PO-xxx` | procurement, finance |
| `SUPPLIER_CONTRACT_APPROVAL` | `RAG-SC-xxx` | procurement, finance |
| `SALES_ORDER_APPROVAL` | `RAG-SO-xxx` | sales |
| `SHIPMENT_APPROVAL` | `RAG-SH-xxx` | sales |
| `BOM_APPROVAL` | `RAG-BM-xxx` | production |
| `ROUTING_APPROVAL` | `RAG-RT-xxx` | production |
| `RAG_WH_RESERVE` | `RAG-WR-xxx` | warehouse |
| `RAG_WH_CONSUME` | `RAG-WC-xxx` | warehouse |
| `RAG_WH_RECEIPT` | `RAG-WT-xxx` | warehouse |

### Эталонные ID документов (`manifest_ids.json`)

| Назначение | UUID | `rag_id` |
|------------|------|----------|
| Заявка на закупку | `e2cdbb12-099c-4a5b-af60-2fc9c05f57b1` | `RAG-PR-001` |
| Заказ на отгрузку | `0f944287-f590-45e5-a8f3-3b1c542cdc3e` | `RAG-SO-001` |
| Договор поставщика | `2f543adf-46c0-4499-8335-16ec889c0aa5` | `RAG-SC-001` |
| Черновик PR (автор `rag_author_proc`) | `b4518951-24e2-4efe-81d5-49ceee9250f0` | — |
| PR на согласовании | `4ba793a8-5adb-4634-b975-66910c950316` | — |
| `type_id` для создания PR | `9a2e4c2b-8f5d-480e-a8fc-dcc120c236d2` | `PURCHASE_REQUEST_APPROVAL` |

---

## Файлы manifest

| Файл | Назначение |
|------|------------|
| `corpus_full.json` | Корпус для индексации: `search_text`, `payload`, `reader_roles` |
| `access_matrix.json` | Список `document_id`, доступных каждому пользователю |
| `manifest_ids.json` | UUID документов и типов для примеров запросов |
| `test_users.json` | Логины, пароли, роли |

### Индексация

1. Индексировать **`search_text`** (или `title` + `payload.description`) из `corpus_full.json`.
2. Для **per-user retrieval** фильтровать по `access_matrix.json` **или** сверять с `GET /api/v1/documents` под JWT пользователя.
3. Внешний ключ в payload: **`rag_id`** (`RAG-PR-001` … `RAG-WT-005`).

---

## Аутентификация (JWT)

Для вызовов API СЭД нужен **access_token** с `tenant_id` тенанта `ragcorp` и realm-ролями пользователя.

**Вариант 1 — Keycloak (через тот же порт 5656):**

```http
POST http://<хост>:5656/realms/industrial-sed/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

grant_type=password
&client_id=auth-service
&client_secret=<CLIENT_SECRET>
&username=rag_finance@ragcorp
&password=CM3L5KExDQQ9oxqQ
```

**Вариант 2 — test login (если включён на стенде):**

```http
POST http://<хост>:5656/api/v1/internal/test/login
X-Test-Secret: <TEST_SECRET>
Content-Type: application/json

{"username":"rag_finance@ragcorp","password":"CM3L5KExDQQ9oxqQ"}
```

**Ответ 200:** `access_token`. Далее:

```http
Authorization: Bearer <access_token>
```

на `http://<хост>:5656/api/v1/...` (СЭД, auth, склад и остальные сервисы маршрутизируются gateway).

---

## Запросы к API СЭД: пользователь → ручка → результат

Переменные для примеров:

```bash
export API_BASE=http://<хост>:5656

PR_ID=e2cdbb12-099c-4a5b-af60-2fc9c05f57b1
SO_ID=0f944287-f590-45e5-a8f3-3b1c542cdc3e
SC_ID=2f543adf-46c0-4499-8335-16ec889c0aa5
DRAFT_ID=b4518951-24e2-4efe-81d5-49ceee9250f0
IN_REVIEW_ID=4ba793a8-5adb-4634-b975-66910c950316
PR_TYPE_ID=9a2e4c2b-8f5d-480e-a8fc-dcc120c236d2
```

Префикс API: `{API_BASE}/api/v1`.

---

### `rag_admin@ragcorp` — полный доступ

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список документов | `GET /documents` | **200**, **50** элементов |
| Только черновики | `GET /documents?status=DRAFT` | **200**, все со `status=DRAFT` |
| Карточка SO | `GET /documents/{SO_ID}` | **200**, `title` содержит `RAG-SO` |
| История PR | `GET /documents/{PR_ID}/history` | **200** |

```bash
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$API_BASE/api/v1/documents" | jq 'length'
# 50
```

---

### `rag_finance@ragcorp` — закупки и договоры (15)

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список | `GET /documents` | **200**, **15** элементов |
| PR | `GET /documents/{PR_ID}` | **200** |
| Договор SC | `GET /documents/{SC_ID}` | **200**, `RAG-SC-001` |
| Чужой SO | `GET /documents/{SO_ID}` | **403** |
| История PR | `GET /documents/{PR_ID}/history` | **200** |
| История SO | `GET /documents/{SO_ID}/history` | **403** |
| Создать PR | `POST /documents` | **403** |

```bash
curl -s -H "Authorization: Bearer $FIN_TOKEN" "$API_BASE/api/v1/documents" | jq 'length'
# 15

curl -s -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $FIN_TOKEN" \
  "$API_BASE/api/v1/documents/$SO_ID"
# 403
```

**Для RAG:** запрос *«договор поставщика RAG-SC»* → в выдаче `RAG-SC-001`…`005`; `RAG-SO-*` **не включать**.

---

### `rag_procurement@ragcorp` — те же 15 закупочных

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список | `GET /documents` | **200**, **15** |
| PR | `GET /documents/{PR_ID}` | **200** |
| SO | `GET /documents/{SO_ID}` | **403** |

---

### `rag_sales@ragcorp` — продажи (10)

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список | `GET /documents` | **200**, **10** |
| SO | `GET /documents/{SO_ID}` | **200** |
| PR | `GET /documents/{PR_ID}` | **403** |

**Для RAG:** *«RAG-SC договор»* под `rag_sales` → пусто; под `rag_finance` → находит SC.

---

### `rag_production@ragcorp` — производство (10)

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список | `GET /documents` | **200**, **10** |
| PR | `GET /documents/{PR_ID}` | **403** |

---

### `rag_warehouse@ragcorp` — склад (15)

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список | `GET /documents` | **200**, **15** |
| PR | `GET /documents/{PR_ID}` | **403** |

---

### `rag_author_proc@ragcorp` — автор закупок

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Свой черновик | `GET /documents/{DRAFT_ID}` | **200** |
| Чужой утверждённый PR | `GET /documents/{PR_ID}` | **403** |
| Создать документ | `POST /documents` с `type_id={PR_TYPE_ID}` | **201** |

```bash
curl -s -X POST -H "Authorization: Bearer $AUTH_PROC_TOKEN" \
  -H "Content-Type: application/json" \
  "$API_BASE/api/v1/documents" \
  -d '{"type_id":"'"$PR_TYPE_ID"'","title":"RAG create","payload":{"description":"..."}}' \
  -w "%{http_code}\n" -o /dev/null
# 201
```

---

### `rag_author_sales@ragcorp` — автор продаж

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Чужой PR | `GET /documents/{PR_ID}` | **403** |
| Создать PR | `POST /documents` (тип закупки) | **403** |

---

### `rag_approver@ragcorp` — согласующий

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Очередь | `GET /tasks` | **200**, есть задача по `{IN_REVIEW_ID}` |
| Документ на согласовании | `GET /documents/{IN_REVIEW_ID}` | **200** |
| Полный реестр | `GET /documents` | **200**, **0** элементов |

---

### `rag_no_access@ragcorp` — нет доступа к реестру

| Действие | Метод и путь | Результат |
|----------|--------------|-----------|
| Список | `GET /documents` | **403** (не пустой массив) |

---

## Сводка для RAG-retrieval

После индексации `corpus_full.json` для каждого пользователя ограничивайте выдачу:

| Пользователь | Допустимо (`rag_id` / типы) | Запрещённый пример |
|--------------|----------------------------|--------------------|
| `rag_finance` | 15 id: `RAG-PR`, `RAG-PO`, `RAG-SC` | `RAG-SO-001` |
| `rag_sales` | 10 id: `RAG-SO`, `RAG-SH` | `RAG-PR-001` |
| `rag_no_access` | ничего (API не отдаёт список) | любой чужой id |

Сверка: `access_matrix.json` → поле `visible_document_ids` для username **или** `GET /api/v1/documents` с JWT этого пользователя.
