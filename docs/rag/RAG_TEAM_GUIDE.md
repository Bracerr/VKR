# RAG — руководство для команды

Корпус из **50 документов СЭД** (10 типов × 5), **11 учётных записей** (включая сервисный аккаунт) с разным ACL и JSON-manifest для индексации и проверки фильтрации по ролям.

  
**Формат логина:** `{username}@ragcorp`  
**Общий пароль всех** `rag_`***:**  Уточнить у администратора

**Базовый URL стенда (один порт для всего API и Keycloak):**

```text
http://85.236.191.21:52565/
```

---

## Учётные записи




| Username           | Логин                      | Пароль                    | Роли                                  | `GET /api/v1/documents`                                    |
| ------------------ | -------------------------- | ------------------------- | ------------------------------------- | ---------------------------------------------------------- |
| `**rag_service**`  | `rag_service@ragcorp`      | Уточнить у администратора | `sed_admin`                           | **200**, **50** — **сервисный аккаунт RAG**, полный доступ |
| `rag_admin`        | `rag_admin@ragcorp`        | тот же                    | `sed_admin`, `sed_author`             | **200**, **50**                                            |
| `rag_finance`      | `rag_finance@ragcorp`      | тот же                    | `doc_read_finance`                    | **200**, **15** (PR, PO, SC)                               |
| `rag_procurement`  | `rag_procurement@ragcorp`  | тот же                    | `doc_read_procurement`                | **200**, **15**                                            |
| `rag_sales`        | `rag_sales@ragcorp`        | тот же                    | `doc_read_sales`                      | **200**, **10** (SO, SH)                                   |
| `rag_production`   | `rag_production@ragcorp`   | тот же                    | `doc_read_production`                 | **200**, **10**                                            |
| `rag_warehouse`    | `rag_warehouse@ragcorp`    | тот же                    | `doc_read_warehouse`                  | **200**, **15**                                            |
| `rag_author_proc`  | `rag_author_proc@ragcorp`  | тот же                    | `sed_author`, `doc_write_procurement` | см. сценарии ниже                                          |
| `rag_author_sales` | `rag_author_sales@ragcorp` | тот же                    | `sed_author`, `doc_write_sales`       | см. сценарии ниже                                          |
| `rag_approver`     | `rag_approver@ragcorp`     | тот же                    | `sed_approver`                        | **200**, **0**; `GET /tasks`                               |
| `rag_no_access`    | `rag_no_access@ragcorp`    | тот же                    | `sed_viewer`                          | **403**                                                    |


### Сервисный аккаунт `rag_service`

Для индексации и batch-запросов RAG-модуля: роль `**sed_admin`** даёт чтение **всех 50** документов тенанта без фильтра `doc_read_`*. 

### Тестовая ручка для обхода авторизации.

```bash
curl -s -X POST "$API_BASE/api/v1/internal/test/login" \
  -H "X-Test-Secret: SECRET" -H "Content-Type: application/json" \
  -d '{"username":"rag_service@ragcorp","password":"PASSWORD"}'
```
Secret & Password уточнить у администратора

### Какие типы документов видит роль


| Код типа                     | Префикс `rag_id` | Кто читает (`doc_read_*`) |
| ---------------------------- | ---------------- | ------------------------- |
| `PURCHASE_REQUEST_APPROVAL`  | `RAG-PR-xxx`     | procurement, finance      |
| `PURCHASE_ORDER_APPROVAL`    | `RAG-PO-xxx`     | procurement, finance      |
| `SUPPLIER_CONTRACT_APPROVAL` | `RAG-SC-xxx`     | procurement, finance      |
| `SALES_ORDER_APPROVAL`       | `RAG-SO-xxx`     | sales                     |
| `SHIPMENT_APPROVAL`          | `RAG-SH-xxx`     | sales                     |
| `BOM_APPROVAL`               | `RAG-BM-xxx`     | production                |
| `ROUTING_APPROVAL`           | `RAG-RT-xxx`     | production                |
| `RAG_WH_RESERVE`             | `RAG-WR-xxx`     | warehouse                 |
| `RAG_WH_CONSUME`             | `RAG-WC-xxx`     | warehouse                 |
| `RAG_WH_RECEIPT`             | `RAG-WT-xxx`     | warehouse                 |

---

## Файлы manifest


| Файл                 | Назначение                                                             |
| -------------------- | ---------------------------------------------------------------------- |
| `corpus_full.json`   | Снимок `GET /internal/rag/corpus`: `text`, `access`, `attachments` |
| `access_matrix.json` | Список `document_id`, доступных каждому пользователю                   |
| `manifest_ids.json`  | UUID документов и типов для примеров запросов                          |
| `test_users.json`    | Логины, пароли, роли                                                   |


### Индексация

Публичный `GET /api/v1/documents` отдаёт **только бизнес-карточку** (`title`, `payload` с заказами, поставщиками, строками и т.д.)

Корпус для RAG и матрица видимости — **закрытая сервисная ручка**:

```http
GET http://<хост>:52565/api/v1/internal/rag/corpus
X-Service-Secret: e2e-service-secret
X-Tenant-Id: ragcorp
```

> **Не путать с test login:** `X-Test-Secret` (`AUTH_TEST_SECRET`, обычно `e2e-test-secret`) — только для `/internal/test/login`.  
> Для corpus нужен `**X-Service-Secret`** = `RAG_CORPUS_SECRET` или `**AUTH_SERVICE_SECRET**` (на прод-стенде `**e2e-service-secret**`).  
> `**Authorization: Bearer` не нужен** — достаточно двух заголовков выше.

**Ответ 200** — только массив `documents`, без `users`, `search_text`, `payload`:

```json
{
  "documents": [
    {
      "document_id": "uuid",
      "text": "Заявка на закупку RAG-PR-001 — …\n\nПодразделение: …\nОбоснование: …\nПозиции:\n  1) …",
      "access": {
        "read_roles": ["doc_read_procurement", "doc_read_finance"],
        "write_roles": ["doc_write_procurement"],
        "approve_roles": ["sed_approver"],
        "admin_roles": ["sed_admin"]
      },
      "attachments": [
        {
          "file_id": "uuid",
          "name": "scan.pdf",
          "content_type": "application/pdf",
          "size_bytes": 12345,
          "url": "http://<хост>:52565/api/v1/internal/rag/documents/<doc_id>/files/<file_id>"
        }
      ]
    }
  ]
}
```

| Поле | Назначение |
|------|------------|
| `text` | Всё, что «написано» в карточке (заголовок + поля payload текстом) — **это и индексирует RAG** |
| `access.read_roles` | Кто может **читать** (realm-роли + логика `sed_admin` / автор в приложении) |
| `access.write_roles` | Кто может **создавать** документы этого типа |
| `access.approve_roles` | Кто может **согласовывать** (`sed_approver`) |
| `access.admin_roles` | Полный доступ (`sed_admin`) |
| `attachments[].url` | Скачать **бинарник** тем же `X-Service-Secret` + `X-Tenant-Id` (GET по `url`) |

```bash
curl -s "http://<хост>:52565/api/v1/internal/rag/corpus" \
  -H "X-Service-Secret: e2e-service-secret" \
  -H "X-Tenant-Id: ragcorp" | jq '.documents | length'
```

Секрет: `RAG_CORPUS_SECRET` или `AUTH_SERVICE_SECRET` (не `AUTH_TEST_SECRET`).  
База URL в ссылках на файлы: `RAG_CORPUS_BASE_URL` / `API_PUBLIC_URL` на стенде.

Публичный `GET /api/v1/documents` — только для UI/проверок ACL; **корпус для RAG — только `/internal/rag/corpus`**.

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
&password=RagTest2026!
```

**Вариант 2 — test login (если включён на стенде):**

```http
POST http://<хост>:52565/api/v1/internal/test/login
X-Test-Secret: e2e-test-secret
Content-Type: application/json

{"username":"rag_service@ragcorp","password":"RagTest2026!"}
```

> **Важно для Postman (test login):** `X-Test-Secret` = `AUTH_TEST_SECRET` (часто `e2e-test-secret`).  
> **Corpus API:** `X-Service-Secret` = `AUTH_SERVICE_SECRET` (на стенде `85.236.191.21` — `**e2e-service-secret`**).  
> При неверном секрете в обоих случаях приходит `401` «неверные учётные данные».

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


| Действие          | Метод и путь                     | Результат                          |
| ----------------- | -------------------------------- | ---------------------------------- |
| Список документов | `GET /documents`                 | **200**, **50** элементов          |
| Только черновики  | `GET /documents?status=DRAFT`    | **200**, все со `status=DRAFT`     |
| Карточка SO       | `GET /documents/{SO_ID}`         | **200**, `title` содержит `RAG-SO` |
| История PR        | `GET /documents/{PR_ID}/history` | **200**                            |


```bash
curl -s -H "Authorization: Bearer $ADMIN_TOKEN" "$API_BASE/api/v1/documents" | jq 'length'
# 50
```

---

### `rag_finance@ragcorp` — закупки и договоры (15)


| Действие   | Метод и путь                     | Результат                 |
| ---------- | -------------------------------- | ------------------------- |
| Список     | `GET /documents`                 | **200**, **15** элементов |
| PR         | `GET /documents/{PR_ID}`         | **200**                   |
| Договор SC | `GET /documents/{SC_ID}`         | **200**, `RAG-SC-001`     |
| Чужой SO   | `GET /documents/{SO_ID}`         | **403**                   |
| История PR | `GET /documents/{PR_ID}/history` | **200**                   |
| История SO | `GET /documents/{SO_ID}/history` | **403**                   |
| Создать PR | `POST /documents`                | **403**                   |


```bash
curl -s -H "Authorization: Bearer $FIN_TOKEN" "$API_BASE/api/v1/documents" | jq 'length'
# 15

curl -s -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $FIN_TOKEN" \
  "$API_BASE/api/v1/documents/$SO_ID"
# 403
```

**Для RAG:** запрос *«договор поставщика RAG-SC»* → в выдаче `RAG-SC-001`…`005`; `RAG-SO-`* **не включать**.

---

### `rag_procurement@ragcorp` — те же 15 закупочных


| Действие | Метод и путь             | Результат       |
| -------- | ------------------------ | --------------- |
| Список   | `GET /documents`         | **200**, **15** |
| PR       | `GET /documents/{PR_ID}` | **200**         |
| SO       | `GET /documents/{SO_ID}` | **403**         |


---

### `rag_sales@ragcorp` — продажи (10)


| Действие | Метод и путь             | Результат       |
| -------- | ------------------------ | --------------- |
| Список   | `GET /documents`         | **200**, **10** |
| SO       | `GET /documents/{SO_ID}` | **200**         |
| PR       | `GET /documents/{PR_ID}` | **403**         |


**Для RAG:** *«RAG-SC договор»* под `rag_sales` → пусто; под `rag_finance` → находит SC.

---

### `rag_production@ragcorp` — производство (10)


| Действие | Метод и путь             | Результат       |
| -------- | ------------------------ | --------------- |
| Список   | `GET /documents`         | **200**, **10** |
| PR       | `GET /documents/{PR_ID}` | **403**         |


---

### `rag_warehouse@ragcorp` — склад (15)


| Действие | Метод и путь             | Результат       |
| -------- | ------------------------ | --------------- |
| Список   | `GET /documents`         | **200**, **15** |
| PR       | `GET /documents/{PR_ID}` | **403**         |


---

### `rag_author_proc@ragcorp` — автор закупок


| Действие              | Метод и путь                               | Результат |
| --------------------- | ------------------------------------------ | --------- |
| Свой черновик         | `GET /documents/{DRAFT_ID}`                | **200**   |
| Чужой утверждённый PR | `GET /documents/{PR_ID}`                   | **403**   |
| Создать документ      | `POST /documents` с `type_id={PR_TYPE_ID}` | **201**   |


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


| Действие   | Метод и путь                    | Результат |
| ---------- | ------------------------------- | --------- |
| Чужой PR   | `GET /documents/{PR_ID}`        | **403**   |
| Создать PR | `POST /documents` (тип закупки) | **403**   |


---

### `rag_approver@ragcorp` — согласующий


| Действие                 | Метод и путь                    | Результат                                |
| ------------------------ | ------------------------------- | ---------------------------------------- |
| Очередь                  | `GET /tasks`                    | **200**, есть задача по `{IN_REVIEW_ID}` |
| Документ на согласовании | `GET /documents/{IN_REVIEW_ID}` | **200**                                  |
| Полный реестр            | `GET /documents`                | **200**, **0** элементов                 |


---

### `rag_no_access@ragcorp` — нет доступа к реестру


| Действие | Метод и путь     | Результат                  |
| -------- | ---------------- | -------------------------- |
| Список   | `GET /documents` | **403** (не пустой массив) |


---

## Сводка для RAG-retrieval

После индексации `corpus_full.json` для каждого пользователя ограничивайте выдачу:


| Пользователь    | Допустимо (`rag_id` / типы)         | Запрещённый пример |
| --------------- | ----------------------------------- | ------------------ |
| `rag_finance`   | 15 id: `RAG-PR`, `RAG-PO`, `RAG-SC` | `RAG-SO-001`       |
| `rag_sales`     | 10 id: `RAG-SO`, `RAG-SH`           | `RAG-PR-001`       |
| `rag_no_access` | ничего (API не отдаёт список)       | любой чужой id     |


Сверка: `access_matrix.json` → поле `visible_document_ids` для username **или** `GET /api/v1/documents` с JWT этого пользователя.