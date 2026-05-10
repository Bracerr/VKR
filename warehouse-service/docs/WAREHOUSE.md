# Warehouse Service — модель и API

## Мультитенантность

Во всех операциях используется **`tenant_id`** из JWT (claim). Cross-tenant запросы невозможны: данные фильтруются по `tenant_code` в БД.

## Роли (realm)

| Роль | Доступ |
|------|--------|
| `warehouse_admin` | CRUD справочников, импорт, все операции |
| `storekeeper` | Операции, резервы, чтение |
| `warehouse_viewer` | Только GET |

## Учёт партий и FEFO

- Товары: `tracking_mode` = `NONE` | `BATCH` | `SERIAL` | `BATCH_AND_SERIAL`.
- Списание без указания партии для `BATCH` / `BATCH_AND_SERIAL` (без перечня серийников) идёт **FEFO**: `expires_at ASC NULLS LAST`, затем `batch id`.
- Явная партия: `batch_id` в теле расхода/перемещения.
- Серийники: расход/перемещение по списку `serial_numbers` (или резерв по `serial_no`).

## Остатки и резервы

Таблица `stock_balances`: `quantity`, `reserved_qty`, `value`. Инвариант: `quantity >= 0`, `reserved_qty <= quantity`.

`available = quantity - reserved_qty` проверяется при резерве и расходе.

Просроченные резервы (`expires_at`) снимаются фоновой задачей (статус `EXPIRED`, возврат `reserved_qty`).

## Основные REST-ручки

Все под `/api/v1` с JWT.

### Справочники (admin — мутации)

- `GET/POST /products`, `GET/PUT/DELETE /products/:id`
- `GET/POST /warehouses`, `PUT/DELETE /warehouses/:id`
- `GET /warehouses/:warehouse_id/bins`, `POST /warehouses/:warehouse_id/bins`, `PUT/DELETE /bins/:id`
- `GET /products/:id/prices`, `POST /products/:id/prices`, `DELETE /prices/:id`
- `GET /batches/:id`
- `GET /serials?product_id=&status=&warehouse_id=`, `GET /serials/:id/history`

### Операции (storekeeper+)

- `POST /operations/receipt` — тело: `warehouse_id`, `bin_id`, `lines[]` (qty, series, expires_at, unit_cost, serial_numbers…)
- `POST /operations/issue` — `warehouse_id`, `bin_id`, `lines[]` (qty или serial_numbers, опционально batch_id)
- `POST /operations/transfer` — from/to склад+ячейка, `lines[]`
- `POST /operations/relocate` — один склад, две ячейки, `lines[]`

### Инвентаризация

- `POST /inventory` — черновик + строки по текущим остаткам (`warehouse_id`, опционально `bin_id`)
- `GET /inventory/:id` — документ + строки
- `PATCH /inventory/lines/:line_id` — факт `counted`
- `POST /inventory/:id/post` — проводка (`INVENTORY_ADJUST`)

### Резервы

- `POST /reservations` — `warehouse_id`, `bin_id`, `product_id`, `qty`, опционально `batch_id` / `serial_no`, `expires_at`
- `GET /reservations`, `GET /reservations/:id`
- `POST /reservations/:id/release`, `POST /reservations/:id/consume` (движение `RESERVE_CONSUMED`)

### Отчёты и остатки (viewer+)

- `GET /balances` — фильтры query: warehouse_id, bin_id, product_id, only_positive, expires_before
- `GET /movements?from=&to=` (RFC3339)
- `GET /reports/stock-on-date?at=`
- `GET /reports/turnover?from=&to=&group_by=`
- `GET /reports/abc?from=&to=&metric=qty|value`
- `GET /reports/expiring?until=` (дата)
- `GET /reports/price-on-date?product_id=&on=&price_type=`
- `GET /reports/average-cost?product_id=&at=`

### Импорт / экспорт

- `POST /import/products` — raw body CSV (первая строка заголовок: `sku,name,unit,tracking_mode`)
- `GET /import/jobs/:id`
- `GET /export/movements.csv?from=&to=`

## Эксплуатация

- Миграции: `migrations/000001_init.up.sql`, опция `run_migrations_on_start` в конфиге.
- Health: `GET /health`, readiness: `GET /ready` (Postgres + Keycloak realm).
- Docker: см. корневой `docker-compose.yaml`.
