# СЭД: модель и API

## Мультитенантность

Во всех запросах к `/api/v1` требуется JWT с claim **`tenant_id`** и ролями realm. Данные изолированы по `tenant_code` в БД.

## Сущности

| Таблица | Назначение |
|---------|------------|
| `workflows` | Маршрут согласования (код уникален в тенанте) |
| `workflow_steps` | Шаги: `order_no`, опционально `parallel_group`, `required_role` **или** `required_user_sub` |
| `document_types` | Тип документа, `warehouse_action` (`NONE`/`RESERVE`/`CONSUME`/`RECEIPT`), `default_workflow_id` |
| `documents` | Экземпляр: статус, автор (`author_sub`), `payload` (JSON для склада), `warehouse_ref` (ответ интеграции) |
| `document_approvals` | Строки согласования по шагам |
| `document_files` | Метаданные файлов в MinIO |
| `document_history` | Аудит действий |

## Статусы документа

`DRAFT` → `IN_REVIEW` → `APPROVED` → `SIGNED`; `CANCELLED`; при отклонении возврат в **`DRAFT`** (записи согласований сбрасываются).

## Workflow

- При **submit** создаются `document_approvals` для всех шагов маршрута типа документа; `current_order_no` = минимальный `order_no` среди шагов.
- **Approve**: учитывается параллельность на текущем `order_no` (все pending на шаге должны быть закрыты, прежде чем перейти к следующему минимальному pending `order_no` или в `APPROVED`).
- **Sign**: только автор; идемпотентность интеграции — если `warehouse_ref` уже заполнен, повторный вызов склада не выполняется; при статусе `SIGNED` повторный sign — no-op.

## REST (`/api/v1`, JWT)

### Справочники (`sed_admin`)

- `GET/POST /document-types`, `GET/PUT/DELETE /document-types/:id`
- `GET/POST /workflows`, `PUT/DELETE /workflows/:id`
- `GET/POST /workflows/:id/steps`, `DELETE /workflow-steps/:id`

### Документы

- `GET /documents`, `GET /documents/:id` — `sed_viewer` и выше
- `POST /documents`, `PATCH /documents/:id`, `POST .../submit|sign|cancel`, вложения — `sed_author` (и `sed_admin`)
- `POST .../approve`, `.../reject`, `GET /tasks` — `sed_approver` (и `sed_admin`)
- `GET /documents/:id/history`
- Файлы: `GET/POST /documents/:id/files`, `GET .../files/:file_id`, `DELETE .../files/:file_id`

### Payload для склада

JSON поля: `warehouse_id`, `default_bin_id`, `lines[]` (`product_id`, `qty`, `bin_id`, …), для consume — `reservation_ids[]`.

## Warehouse-service

Сервис вызывает `warehouse_base_url` с заголовками:

- `X-Service-Secret` — общий секрет
- `X-Tenant-Id` — код тенанта из JWT

На стороне warehouse включён альтернативный путь аутентификации: при валидном секрете выставляются synthetic claims с ролью `warehouse_admin`.

## Интеграция с production-service (MES)

Для согласования BOM и техкарт в тенанте создаются типы документов с **`warehouse_action: NONE`** (например коды `BOM_APPROVAL`, `ROUTING_APPROVAL`), к ним привязывается маршрут согласования.

После успешного **`POST /documents/:id/sign`** sed-service может вызвать **production-service** (если заданы в конфиге `production_callback_url` и `production_callback_secret`):

- `POST {production_callback_url}/api/v1/internal/sed-events`
- Заголовок **`X-Service-Secret`**: тот же секрет, что **`sed_callback_verify_secret`** у production-service.
- Тело: `{ "event": "DOCUMENT_SIGNED", "tenant_code": "<tenant>", "document_id": "<uuid>", "document_type_code": "<code>" }`.

Production находит BOM или routing по `sed_document_id` и переводит запись в статус **`APPROVED`**.

## Интеграция с procurement-service (закупки)

Для согласования PR/PO в тенанте создаются типы документов с **`warehouse_action: NONE`** (например коды `PURCHASE_REQUEST_APPROVAL`, `PURCHASE_ORDER_APPROVAL`, опционально `SUPPLIER_CONTRACT_APPROVAL`).

После успешного **`POST /documents/:id/sign`** sed-service может вызвать **procurement-service** (если заданы в конфиге `procurement_callback_url` и `procurement_callback_secret`):

- `POST {procurement_callback_url}/api/v1/internal/sed-events`
- Заголовок **`X-Service-Secret`**: тот же секрет, что **`sed_callback_verify_secret`** у procurement-service.
- Тело: `{ "event": "DOCUMENT_SIGNED", "tenant_code": "<tenant>", "document_id": "<uuid>", "document_type_code": "<code>" }`.

Procurement находит PR/PO по `sed_document_id` и переводит запись в статус **`APPROVED`**.
