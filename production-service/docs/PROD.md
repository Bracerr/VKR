# Производственный учёт (production-service)

## Сущности

- **workcenters** — рабочие центры (станок/участок).
- **scrap_reasons** — справочник причин брака.
- **boms** / **bom_lines** — версионируемые спецификации; статусы `DRAFT` → `SUBMITTED` (в СЭД) → `APPROVED` / `ARCHIVED`.
- **routings** / **routing_operations** — техкарты и операции с привязкой к `workcenter`.
- **production_orders** — заказы; статусы `PLANNED` → `RELEASED` → `IN_PROGRESS` → `COMPLETED` / `CANCELLED`; поле `reservations` (JSON): связь `bom_line_id` ↔ `reservation_id` склада.
- **production_order_operations** — snapshot операций маршрута на момент release.
- **shift_tasks** — сменные задания на операцию заказа.
- **production_reports** — журнал отчётов по операциям.
- **production_history** — аудит.

## REST (базовый префикс `/api/v1`)

Внутренний callback (без JWT, только `X-Service-Secret`):

- `POST /internal/sed-events` — тело `{ "event":"DOCUMENT_SIGNED", "tenant_code":"...", "document_id":"..." }`.

Аутентификация: `Authorization: Bearer <JWT>`.

Справочники и BOM/маршруты (см. Swagger после `make swagger`): workcenters, scrap-reasons, boms, routings, submit в СЭД.

Заказы: `POST /orders`, `POST /orders/:id/release|cancel|complete`.

Операции заказа: `POST /orders/:id/operations/:op_id/start|report|finish` (в URL заказа — параметр `:id`, как у `GET /orders/:id`).

Смены: `GET /shift-tasks`, `POST /shift-tasks`, `GET /me/shift-tasks`, `DELETE /shift-tasks/:id`.

## Согласование через СЭД

1. Технолог создаёт типы документов в sed (`BOM_APPROVAL`, `ROUTING_APPROVAL`) с маршрутом и `warehouse_action: NONE`.
2. `POST /boms/:id/submit` / `POST /routings/:id/submit` создают документ в sed и вызывают submit (Bearer пользователя).
3. После полного согласования автор подписывает документ (`sign` в sed).
4. sed вызывает production `DOCUMENT_SIGNED` → BOM/маршрут переходят в `APPROVED`.

## Склад

Используются существующие API warehouse с заголовками `X-Service-Secret` и `X-Tenant-Id` (как в sed-service).
