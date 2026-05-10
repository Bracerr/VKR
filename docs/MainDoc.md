# Гайд для фронтенда: бизнес-логика проекта

Документ для фронтенд-разработчика: **какие бизнес-контуры есть**, **какие роли/права**, **какие статусы и сценарии**, и **куда смотреть детали** (Swagger и существующие docs).

## 1) Сервисы и точки входа

| Контур | Сервис | Базовый URL (dev по умолчанию) | Swagger |
|---|---|---|---|
| Авторизация/пользователи | `auth-service` | `http://localhost:8080` | `GET /swagger/index.html` |
| Склад | `warehouse-service` | `http://localhost:8090` | `GET /swagger/index.html` |
| СЭД (согласование/подпись, вложения) | `sed-service` | `http://localhost:8091` | `GET /swagger/index.html` |
| Производство (MES) | `production-service` | `http://localhost:8092` | (в MVP ориентируемся на `docs/PROD.md`) |
| Закупки | `procurement-service` | `http://localhost:8093` | (в MVP ориентируемся на `docs/PROC.md`) |

Общее:

- **Мультитенантность**: все бизнес-сервисы читают `tenant_id` из JWT и фильтруют данные по `tenant_code`.
- **Авторизация**: Bearer JWT (из Keycloak через `auth-service`).

## 2) Роли и права (для UI)

Полный список и иерархия ролей описаны в `auth-service/docs/FLOW.md`.

### Базовые роли управления

- **`super_admin`**: создаёт тенанты и первого `ent_admin`.
- **`ent_admin`**: управляет пользователями и назначает роли внутри своего тенанта.

### Склад (warehouse)

- `warehouse_admin`, `storekeeper`, `warehouse_viewer` — см. `warehouse-service/docs/WAREHOUSE.md`.

### СЭД (sed)

- `sed_admin`, `sed_author`, `sed_approver`, `sed_viewer` — см. `sed-service/docs/SED.md`.

### Производство (prod)

- `prod_admin`, `prod_technologist`, `prod_planner`, `prod_master`, `prod_worker`, `prod_viewer` — см. `production-service/docs/PROD.md` и `auth-service/docs/FLOW.md`.

### Закупки (proc)

- `proc_admin`, `proc_buyer`, `proc_viewer` — см. `procurement-service/docs/PROC.md` и `auth-service/docs/FLOW.md`.

## 3) Авторизация в SPA (как логиниться)

Подробно: `auth-service/docs/FLOW.md`.

Коротко для фронта:

- **Login**: `GET /api/v1/auth/login?return_to=/...` → редирект на Keycloak (OIDC + PKCE).
- **Callback**: `GET /api/v1/auth/callback` — ставит httpOnly cookies.
- **Me**: `GET /api/v1/auth/me` (JWT middleware) — информация о текущем пользователе.
- **Refresh**: `POST /api/v1/auth/refresh` (cookies) → `204`.
- **Logout**: `POST /api/v1/auth/logout` → возвращает `end_session_url` (для завершения SSO).

Важно: для browser запросов обычно нужны cookies → `credentials: "include"`.

## 4) СЭД (sed-service): согласование/подпись как базовый механизм

Док: `sed-service/docs/SED.md`.

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

Док: `warehouse-service/docs/WAREHOUSE.md` + Swagger.

### Что умеет склад (MVP)

- **Справочники**: товары (`products`), склады (`warehouses`), ячейки (`bins`), цены.
- **Операции**: приход/расход/перемещение/перекладка, инвентаризация.
- **Резервы**: create/release/consume; инвариант `reserved_qty <= quantity`.
- **Отчёты**: `GET /balances`, `GET /movements`, отчёты по датам.

### FEFO/партии/серийники

См. раздел “Учёт партий и FEFO” в `warehouse-service/docs/WAREHOUSE.md`.

## 6) Производство (production-service): BOM/маршруты/заказы/операции/смены

Док: `production-service/docs/PROD.md`.

### Сущности и статусы (главное для UI)

- **BOM**: `DRAFT` → `SUBMITTED` → (после подписи в СЭД) `APPROVED` → `ARCHIVED`.
- **Routing**: аналогично BOM.
- **Production order**: `PLANNED` → `RELEASED` → `IN_PROGRESS` → `COMPLETED` (или `CANCELLED`).
- **Shift tasks**: сменные задания по операциям заказов.

### Интеграция с СЭД

- Production создаёт документ в SED при submit BOM/маршрута.
- После `SIGNED` в SED → callback в production (`/api/v1/internal/sed-events`) → BOM/маршрут становятся `APPROVED`.

### Интеграция со складом

Production использует операции склада (резервы/списания/приход готовой продукции) через service-secret паттерн (детали см. `production-service/docs/PROD.md`).

## 7) Закупки (procurement-service): PR → PO → Receipt (в склад)

Док: `procurement-service/docs/PROC.md`.

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

Seed типов документов закупок для SED: `procurement-service/scripts/example_document_types.sql`.

## 8) Сквозные пользовательские сценарии (что “может система”)

### (A) Онбординг тенанта и пользователей

См. `auth-service/docs/FLOW.md`:

- super_admin создаёт tenant
- super_admin создаёт `ent_admin`
- `ent_admin` создаёт пользователей и назначает роли доменов (warehouse/sed/prod/proc)

### (B) СЭД-документы с вложениями и согласованием

См. `sed-service/docs/SED.md` + Swagger.

### (C) Складской учёт

См. `warehouse-service/docs/WAREHOUSE.md` + Swagger.

### (D) Производственный контур

См. `production-service/docs/PROD.md`.

### (E) Закупочный контур

См. `procurement-service/docs/PROC.md`.

## 9) Где смотреть детали API

- `auth-service` Swagger: `GET /swagger/index.html`
- `warehouse-service` Swagger: `GET /swagger/index.html`
- `sed-service` Swagger: `GET /swagger/index.html`
- Доп. docs:
  - `auth-service/docs/FLOW.md`
  - `warehouse-service/docs/WAREHOUSE.md`
  - `sed-service/docs/SED.md`
  - `production-service/docs/PROD.md`
  - `procurement-service/docs/PROC.md`

