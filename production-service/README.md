# production-service

Микросервис производственного учёта (MES): спецификации (BOM), техкарты (routing), рабочие центры, производственные заказы, сменные задания, факт выработки; интеграция со **warehouse-service** (резервы/списание/приход ГП) и **sed-service** (согласование BOM/маршрутов, callback после подписи).

## Стек

- Go 1.22, Gin, PostgreSQL (pgx/v5), slog, golang-migrate, shopspring/decimal
- JWT Keycloak (JWKS), роли `prod_*` (см. auth-service)
- Swagger: `make swagger` → `docs/` (после генерации)

## Быстрый старт

1. PostgreSQL на порту **5435**, Keycloak, auth-service, warehouse-service, sed-service (см. `configs/local.yaml`).
2. Скопировать `.env.example`, задать `WAREHOUSE_SERVICE_SECRET`, `SED_CALLBACK_VERIFY_SECRET` (тот же секрет, что `PRODUCTION_CALLBACK_SECRET` в sed для callback).
3. `make migrate-up` или `run_migrations_on_start: true`.
4. `CONFIG_PATH=./configs/local.yaml make run` — сервис **:8092**.

```bash
CONFIG_PATH=./configs/local.yaml make run
```

## Docker

```bash
make up   # prod-db :5435 + production-service :8092
```

## Документация

- [docs/PROD.md](docs/PROD.md) — модель данных, REST, workflow, интеграции
- [docs/TESTING.md](docs/TESTING.md) — юнит-тесты и e2e (общий compose с sed-service)

## Роли Keycloak

`prod_admin`, `prod_technologist`, `prod_planner`, `prod_master`, `prod_worker`, `prod_qc`, `prod_viewer` — задаются в **auth-service** (`RealmRoles`, bootstrap ролей).

## Callback СЭД → production

После `POST /documents/:id/sign` sed-service вызывает `POST /api/v1/internal/sed-events` с заголовком `X-Service-Secret`, если заданы `production_callback_url` и `production_callback_secret` в конфиге sed.

## Интеграция со складом

При `POST /orders/:id/release` создаются резервы по строкам BOM; при `POST .../operations/:op_id/finish` — consume резервов по строкам с тем же `op_no`; при `POST /orders/:id/complete` — приход готовой продукции (`operations/receipt`).
