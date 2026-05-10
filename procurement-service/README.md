# procurement-service

`procurement-service` — контур закупок (MVP): **PR → PO → приемка по PO** с интеграцией со складом (`warehouse-service`) и согласованием/подписью в СЭД (`sed-service`).

## Быстрый старт

### Локально (Postgres)

```bash
cd procurement-service
cp .env.example .env
make up
```

Сервис поднимется на `:8093`, Postgres на `:5436`.

### Миграции

```bash
cd procurement-service
make migrate-up
```

## Конфигурация

См. `configs/local.yaml`, `configs/docker.yaml` и `.env.example`.

Ключевые переменные:

- `DB_DSN`
- `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`
- `WAREHOUSE_BASE_URL`, `WAREHOUSE_SERVICE_SECRET`
- `SED_BASE_URL`
- `SED_CALLBACK_VERIFY_SECRET` — секрет для `POST /api/v1/internal/sed-events` (выставляется в `sed-service` как `PROCUREMENT_CALLBACK_SECRET`)

## Документация

- `docs/PROC.md`
- `docs/TESTING.md`
- Seed типов документов для SED: `scripts/example_document_types.sql`

