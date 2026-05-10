# Procurement-service (закупки)

## Скоуп MVP

- Поставщики (`suppliers`)
- Заявки на закупку PR (`purchase_requests` + `purchase_request_lines`)
- Заказы поставщику PO (`purchase_orders` + `purchase_order_lines`)
- Приемка по PO (`receipts`) + складской приход в `warehouse-service`
- Аудит (`procurement_history`)

## Статусы

### PR

- `DRAFT` → `SUBMITTED` → `APPROVED`
- `CANCELLED`

### PO

- `DRAFT` → `SUBMITTED` → `APPROVED` → `RELEASED` → `RECEIVED`
- `CANCELLED`

## Интеграции

### warehouse-service

Приемка по PO вызывает:

- `POST /api/v1/operations/receipt`
- Заголовки: `X-Service-Secret`, `X-Tenant-Id`, `Idempotency-Key`

Идемпотентность — по ключу `po-receipt-<po_id>`.

### sed-service

Procurement создаёт документ согласования от имени пользователя (Bearer) и вызывает submit:

- `POST /api/v1/documents`
- `POST /api/v1/documents/:id/submit`

После подписи (`SIGNED`) sed-service делает callback:

- `POST /api/v1/internal/sed-events`
- Заголовок: `X-Service-Secret` = `SED_CALLBACK_VERIFY_SECRET` procurement
- Тело: `{ "event":"DOCUMENT_SIGNED", "tenant_code":"...", "document_id":"...", "document_type_code":"..." }`

## API (MVP)

### Health

- `GET /health`
- `GET /ready`

### Suppliers

- `GET /api/v1/suppliers`
- `POST /api/v1/suppliers`
- `PUT /api/v1/suppliers/:id`
- `DELETE /api/v1/suppliers/:id`

### PR

- `GET /api/v1/purchase-requests?status=`
- `GET /api/v1/purchase-requests/:id`
- `POST /api/v1/purchase-requests`
- `POST /api/v1/purchase-requests/:id/lines`
- `POST /api/v1/purchase-requests/:id/submit`
- `POST /api/v1/purchase-requests/:id/cancel`

### PO

- `GET /api/v1/purchase-orders?status=`
- `GET /api/v1/purchase-orders/:id`
- `POST /api/v1/purchase-orders`
- `POST /api/v1/purchase-orders/from-pr/:id`
- `POST /api/v1/purchase-orders/:id/lines`
- `POST /api/v1/purchase-orders/:id/submit`
- `POST /api/v1/purchase-orders/:id/release`
- `POST /api/v1/purchase-orders/:id/cancel`
- `POST /api/v1/purchase-orders/:id/receive`

## RBAC

- Чтение: `proc_viewer` и выше
- Создание/ведение/submit/release/receive: `proc_buyer` (или `proc_admin`)

