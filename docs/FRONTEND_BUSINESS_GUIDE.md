# Гайд для фронтенда: бизнес-логика проекта

Документ для фронтенд-разработчика: **какие бизнес-контуры есть**, **какие роли/права**, **какие статусы и сценарии**, и **куда смотреть детали** (Swagger и существующие docs).

## 1) Сервисы и точки входа

| Контур | Сервис | Базовый URL (dev по умолчанию) | Swagger |
|---|---|---|---|
| Авторизация/пользователи | `auth-service` | `http://localhost:8080` | [`auth-service/docs/swagger.yaml`](../auth-service/docs/swagger.yaml) (или [`swagger.json`](../auth-service/docs/swagger.json)) |
| Склад | `warehouse-service` | `http://localhost:8090` | [`warehouse-service/docs/swagger.yaml`](../warehouse-service/docs/swagger.yaml) (или [`swagger.json`](../warehouse-service/docs/swagger.json)) |
| СЭД (согласование/подпись, вложения) | `sed-service` | `http://localhost:8091` | [`sed-service/docs/swagger.yaml`](../sed-service/docs/swagger.yaml) (или [`swagger.json`](../sed-service/docs/swagger.json)) |
| Производство (MES) | `production-service` | `http://localhost:8092` | (в MVP ориентируемся на `docs/PROD.md`) |
| Закупки | `procurement-service` | `http://localhost:8093` | (в MVP ориентируемся на `docs/PROC.md`) |
| Продажи и отгрузки | `sales-service` | `http://localhost:8094` | (в MVP ориентируемся на `docs/SALES.md`) |
| Прослеживаемость | `traceability-service` | `http://localhost:8095` | (в MVP ориентируемся на `docs/TRACE.md`) |

Общее:

- **Мультитенантность**: все бизнес-сервисы читают `tenant_id` из JWT и фильтруют данные по `tenant_code`.
- **Авторизация**: Bearer JWT (из Keycloak через `auth-service`).

## 2) Роли и права (для UI)

Полный список и иерархия ролей описаны в [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md).

### Базовые роли управления

- **`super_admin`**: создаёт тенанты и первого `ent_admin`.
- **`ent_admin`**: управляет пользователями и назначает роли внутри своего тенанта.

### Склад (warehouse)

- `warehouse_admin`, `storekeeper`, `warehouse_viewer` — см. [`warehouse-service/docs/WAREHOUSE.md`](../warehouse-service/docs/WAREHOUSE.md).

### СЭД (sed)

- `sed_admin`, `sed_author`, `sed_approver`, `sed_viewer` — см. [`sed-service/docs/SED.md`](../sed-service/docs/SED.md).

### Производство (prod)

- `prod_admin`, `prod_technologist`, `prod_planner`, `prod_master`, `prod_worker`, `prod_viewer` — см. [`production-service/docs/PROD.md`](../production-service/docs/PROD.md) и [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md).

### Закупки (proc)

- `proc_admin`, `proc_buyer`, `proc_viewer` — см. [`procurement-service/docs/PROC.md`](../procurement-service/docs/PROC.md) и [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md).

### Продажи (sales)

- `sales_admin`, `sales_manager`, `sales_viewer` — см. [`sales-service/docs/SALES.md`](../sales-service/docs/SALES.md) и [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md).

## 3) Авторизация в SPA (как логиниться)

Подробно: [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md).

Коротко для фронта:

- **Login**: `GET /api/v1/auth/login?return_to=/...` → редирект на Keycloak (OIDC + PKCE).
- **Callback**: `GET /api/v1/auth/callback` — ставит httpOnly cookies.
- **Me**: `GET /api/v1/auth/me` (JWT middleware) — информация о текущем пользователе.
- **Refresh**: `POST /api/v1/auth/refresh` (cookies) → `204`.
- **Logout**: `POST /api/v1/auth/logout` → возвращает `end_session_url` (для завершения SSO).

Важно: для browser запросов обычно нужны cookies → `credentials: "include"`.

## 4) СЭД (sed-service): согласование/подпись как базовый механизм

Док: [`sed-service/docs/SED.md`](../sed-service/docs/SED.md).

### Модель

- **workflows** и **workflow_steps**: маршрут согласования (шаги с `required_role` или `required_user_sub`, возможна параллельность).
- **document_types**: тип документа с `warehouse_action` (`NONE/RESERVE/CONSUME/RECEIPT`) и `default_workflow_id`.
- **documents**: документ со статусом и `payload` (JSON), `warehouse_ref` (результат интеграции).

### Статусы документа

`DRAFT` → `IN_REVIEW` → `APPROVED` → `SIGNED`; плюс `CANCELLED`.

### Ключевой UX-паттерн

- Автор создаёт документ (черновик), заполняет payload, прикладывает файлы.
- Отправляет на согласование (submit).
- Согласующие видят задачи (`GET /tasks`) и делают approve/reject.
- После `APPROVED` автор делает `sign`.

## 5) Склад (warehouse-service): справочники, операции, остатки

Док: [`warehouse-service/docs/WAREHOUSE.md`](../warehouse-service/docs/WAREHOUSE.md) + Swagger ([`warehouse-service/docs/swagger.yaml`](../warehouse-service/docs/swagger.yaml)).

### Что умеет склад (MVP)

- **Справочники**: товары (`products`), склады (`warehouses`), ячейки (`bins`), цены.
- **Операции**: приход/расход/перемещение/перекладка, инвентаризация.
- **Резервы**: create/release/consume; инвариант `reserved_qty <= quantity`.
- **Отчёты**: `GET /balances`, `GET /movements`, отчёты по датам.

### FEFO/партии/серийники

См. раздел “Учёт партий и FEFO” в [`warehouse-service/docs/WAREHOUSE.md`](../warehouse-service/docs/WAREHOUSE.md).

## 6) Производство (production-service): BOM/маршруты/заказы/операции/смены

Док: [`production-service/docs/PROD.md`](../production-service/docs/PROD.md).

### Сущности и статусы (главное для UI)

- **BOM**: `DRAFT` → `SUBMITTED` → (после подписи в СЭД) `APPROVED` → `ARCHIVED`.
- **Routing**: аналогично BOM.
- **Production order**: `PLANNED` → `RELEASED` → `IN_PROGRESS` → `COMPLETED` (или `CANCELLED`).
- **Shift tasks**: сменные задания по операциям заказов.

### Интеграция с СЭД

- Production создаёт документ в SED при submit BOM/маршрута.
- После `SIGNED` в SED → callback в production (`/api/v1/internal/sed-events`) → BOM/маршрут становятся `APPROVED`.

### Интеграция со складом

Production использует операции склада (резервы/списания/приход готовой продукции) через service-secret паттерн (детали см. [`production-service/docs/PROD.md`](../production-service/docs/PROD.md)).

## 7) Закупки (procurement-service): PR → PO → Receipt (в склад)

Док: [`procurement-service/docs/PROC.md`](../procurement-service/docs/PROC.md).

### Сущности и статусы

- **Supplier**
- **PR**: `DRAFT` → `SUBMITTED` → `APPROVED` или `CANCELLED`
- **PO**: `DRAFT` → `SUBMITTED` → `APPROVED` → `RELEASED` → `RECEIVED` или `CANCELLED`
- **Receipt**: факт приемки (`POSTED`) + ссылка на складской документ прихода

### Согласование через СЭД (обязательный шаг)

- PR/PO по кнопке submit создают документ в SED и отправляют на маршрут.
- После подписи в SED (`SIGNED`) → callback в procurement (`POST /api/v1/internal/sed-events`) → PR/PO становятся `APPROVED`.
- После `PO.APPROVED` закупщик делает `release`, затем `receive`.

### Приёмка

`POST /api/v1/purchase-orders/:id/receive` вызывает складской приход (`warehouse-service /operations/receipt`) с идемпотентностью.

Seed типов документов закупок для SED: [`procurement-service/scripts/example_document_types.sql`](../procurement-service/scripts/example_document_types.sql).

## 8) Продажи и отгрузки (sales-service): SO → Reserve → Ship

Док: [`sales-service/docs/SALES.md`](../sales-service/docs/SALES.md).

### Сущности и статусы

- **Customer**
- **SO**: `DRAFT` → `SUBMITTED` → `APPROVED` → `RELEASED` → `SHIPPED` (или `CANCELLED`)
- **Shipment**: факт отгрузки (`POSTED`) + ссылка на складской документ

### Согласование через СЭД (обязательный шаг)

- `POST /api/v1/sales-orders/:id/submit` создаёт документ в SED и отправляет на маршрут.
- После подписи в SED (`SIGNED`) → callback в sales (`POST /api/v1/internal/sed-events`) → SO становится `APPROVED`.
- Далее менеджер делает `release`, затем `reserve`, затем `ship`.

### Резерв и отгрузка

- `reserve` создаёт резервы в `warehouse-service` и сохраняет `reservation_ids`.
- `ship` списывает товар со склада **по reservation_ids** (через `warehouse-service` операция issue-from-reservations), сохраняет `warehouse_document_id`, SO становится `SHIPPED`.

Seed типов документов продаж для SED: [`sales-service/scripts/example_document_types.sql`](../sales-service/scripts/example_document_types.sql).

## 9) Прослеживаемость (traceability-service): поиск + граф цепочки

Док: [`traceability-service/docs/TRACE.md`](../traceability-service/docs/TRACE.md).

### Что это даёт UI

- **Поиск якоря**: серийник/партия → набор найденных узлов (якорей) для выбора пользователем.
- **Граф цепочки**: получаем `{nodes, edges}` для визуализации связей “документы/движения/бизнес-документы”.

### Основные эндпоинты

- `GET /api/v1/trace/search?serial_no=&batch_id=&product_id=&from=&to=`
- `GET /api/v1/trace/graph?anchor_type=&anchor_id=&from=&to=&depth=`

### Откуда берутся связи

- **Warehouse → Trace**: `warehouse-service` после проведения складских операций отправляет `DocumentPosted` (с batch/serial).
- **Domains → Trace**:
  - `sales-service` после `ship` отправляет связь `SO → warehouse_document_id`
  - `procurement-service` после `receive` отправляет связь `PO → warehouse_document_id`
  - `production-service` после `complete` (приход ГП) отправляет связь `PROD_ORDER → warehouse_document_id`

### E2E сценарий

См. `pytest` сценарии:

- [`sed-service/e2e_tests/tests/test_traceability_smoke.py`](../sed-service/e2e_tests/tests/test_traceability_smoke.py)
- [`sed-service/e2e_tests/tests/test_traceability_flow.py`](../sed-service/e2e_tests/tests/test_traceability_flow.py)

## 9) Сквозные пользовательские сценарии (что “может система”)

### (A) Онбординг тенанта и пользователей

См. [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md):

- super_admin создаёт tenant
- super_admin создаёт `ent_admin`
- `ent_admin` создаёт пользователей и назначает роли доменов (warehouse/sed/prod/proc)

### (B) СЭД-документы с вложениями и согласованием

См. [`sed-service/docs/SED.md`](../sed-service/docs/SED.md) + Swagger ([`sed-service/docs/swagger.yaml`](../sed-service/docs/swagger.yaml)).

### (C) Складской учёт

См. [`warehouse-service/docs/WAREHOUSE.md`](../warehouse-service/docs/WAREHOUSE.md) + Swagger ([`warehouse-service/docs/swagger.yaml`](../warehouse-service/docs/swagger.yaml)).

### (D) Производственный контур

См. [`production-service/docs/PROD.md`](../production-service/docs/PROD.md).

### (E) Закупочный контур

См. [`procurement-service/docs/PROC.md`](../procurement-service/docs/PROC.md).

### (F) Продажи и отгрузка

См. [`sales-service/docs/SALES.md`](../sales-service/docs/SALES.md) и e2e сценарий: [`sed-service/e2e_tests/tests/test_sales_flow.py`](../sed-service/e2e_tests/tests/test_sales_flow.py).

## 10) Где смотреть детали API (в репозитории)

- `auth-service` Swagger: [`auth-service/docs/swagger.yaml`](../auth-service/docs/swagger.yaml) (или [`swagger.json`](../auth-service/docs/swagger.json))
- `warehouse-service` Swagger: [`warehouse-service/docs/swagger.yaml`](../warehouse-service/docs/swagger.yaml) (или [`swagger.json`](../warehouse-service/docs/swagger.json))
- `sed-service` Swagger: [`sed-service/docs/swagger.yaml`](../sed-service/docs/swagger.yaml) (или [`swagger.json`](../sed-service/docs/swagger.json))
- Доп. docs:
  - [`auth-service/docs/FLOW.md`](../auth-service/docs/FLOW.md)
  - [`warehouse-service/docs/WAREHOUSE.md`](../warehouse-service/docs/WAREHOUSE.md)
  - [`sed-service/docs/SED.md`](../sed-service/docs/SED.md)
  - [`production-service/docs/PROD.md`](../production-service/docs/PROD.md)
  - [`procurement-service/docs/PROC.md`](../procurement-service/docs/PROC.md)
  - [`sales-service/docs/SALES.md`](../sales-service/docs/SALES.md)
  - [`traceability-service/docs/TRACE.md`](../traceability-service/docs/TRACE.md)

