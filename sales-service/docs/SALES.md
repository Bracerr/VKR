# Sales-service (SO → Reserve → Ship)

## Сущности (MVP)

- **Customer** (`customers`): контрагент/клиент.
- **Sales Order** (`sales_orders`): заказ клиента (SO) + статус.
- **Sales Order Lines** (`sales_order_lines`): строки заказа.
- **Shipment** (`shipments`): факт отгрузки (привязка к складскому документу).
- **Sales History** (`sales_history`): аудит действий.

## Статусы SO

- **`DRAFT`**: черновик (можно добавлять строки).
- **`SUBMITTED`**: отправлен на согласование в `sed-service`.
- **`APPROVED`**: подтверждён через callback из `sed-service` после подписи документа.
- **`RELEASED`**: разрешён к резервированию/отгрузке.
- **`SHIPPED`**: отгружен (создан складской документ списания).
- **`CANCELLED`**: отменён.

Инварианты:

- Мутации строк разрешены только в `DRAFT`.
- `reserve`/`ship` разрешены только в `RELEASED`.

## Интеграции

### SED (согласование)

- `sales-service` создаёт документ в `sed-service` от имени пользователя и вызывает submit.
- После подписи `sed-service` вызывает internal callback:
  - `POST /api/v1/internal/sed-events` в `sales-service`
  - заголовок `X-Service-Secret` (тот же секрет, что `sed_callback_verify_secret` у `sales-service`)

### Warehouse (резерв и отгрузка)

- Резерв создаётся через `POST /api/v1/reservations` (service-secret + tenant header).
- Отгрузка делается **по reservation_ids** через `POST /api/v1/operations/issue-from-reservations` (создаёт `documents` типа `ISSUE` и движения `ISSUE`).

## API (MVP)

```text
GET /health, /ready

# customers
GET/POST /api/v1/customers
PUT/DELETE /api/v1/customers/:id

# SO
GET  /api/v1/sales-orders?status=
GET  /api/v1/sales-orders/:id
POST /api/v1/sales-orders
POST /api/v1/sales-orders/:id/lines
POST /api/v1/sales-orders/:id/submit
POST /api/v1/sales-orders/:id/release
POST /api/v1/sales-orders/:id/cancel
POST /api/v1/sales-orders/:id/reserve
POST /api/v1/sales-orders/:id/ship

# internal
POST /api/v1/internal/sed-events
```

RBAC:

- чтение: `sales_viewer` и выше
- операции/CRUD: `sales_manager` и `sales_admin`

