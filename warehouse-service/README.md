# warehouse-service

Микросервис склада: справочники, приход/расход (FEFO), перемещение и перекладка между ячейками, инвентаризация, резервы, остатки, отчёты, импорт CSV товаров.

- **Стек:** Go 1.22, Gin, PostgreSQL (pgx), JWT Keycloak (JWKS), slog.
- **Конфиг:** `configs/local.yaml`, секреты через `DB_DSN` и переменные окружения (см. `internal/config`).
- **Документация:** [docs/WAREHOUSE.md](docs/WAREHOUSE.md), тестирование — [docs/TESTING.md](docs/TESTING.md).

## Быстрый старт

```bash
make up          # Postgres :5433 + сервис :8090
make run         # локально с CONFIG_PATH=./configs/local.yaml
```

Роли Keycloak: `warehouse_admin`, `storekeeper`, `warehouse_viewer` (создаются в **auth-service**, см. `auth-service/docs/FLOW.md`).

## API

Базовый префикс: `/api/v1`, заголовок `Authorization: Bearer <access JWT>`.

Основные группы:

- Справочники: `GET/POST /products`, `PUT/DELETE /products/:id`, склады и ячейки, цены.
- Операции: `POST /operations/receipt`, `/operations/issue`, `/operations/transfer`, `/operations/relocate`.
- Инвентаризация: `POST /inventory`, `PATCH /inventory/lines/:line_id`, `POST /inventory/:id/post`.
- Резервы: `POST /reservations`, `POST /reservations/:id/release|consume`.
- Отчёты: `/balances`, `/movements`, `/reports/*`.
- Импорт: `POST /import/products` (тело — CSV), статус `GET /import/jobs/:id`.

## Swagger

```bash
make swagger
```

После генерации: `docs/swagger.json` (подключение `gin-swagger` при необходимости добавьте в `router`).
