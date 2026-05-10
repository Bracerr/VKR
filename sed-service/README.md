# sed-service

Электронный документооборот (СЭД): маршруты согласования, вложения в **MinIO**, интеграция со **warehouse-service** при подписании документа.

## Быстрый старт (локально)

1. Поднять PostgreSQL (порт **5434**), MinIO (**9000/9001**), Keycloak, auth-service и warehouse-service (см. `configs/local.yaml`).
2. Скопировать `.env.example` при необходимости; задать `WAREHOUSE_SERVICE_SECRET` и тот же секрет в warehouse (`SERVICE_SECRET` / `service_secret`).
3. `make migrate-up` или запуск с `run_migrations_on_start: true`.
4. `make run` — сервис слушает **:8091**.

```bash
CONFIG_PATH=./configs/local.yaml make run
```

## Docker

```bash
make up   # docker compose в корне sed-service
```

## Документация

- [docs/SED.md](docs/SED.md) — модель данных, REST, workflow, склад.
- [docs/TESTING.md](docs/TESTING.md) — тесты и e2e.
- Swagger: `make swagger`, UI при запущенном сервисе: `http://localhost:8091/swagger/index.html`.

## Роли Keycloak (realm)

`sed_admin`, `sed_author`, `sed_approver`, `sed_viewer` — описаны в **auth-service** (`RealmRoles` + `docs/FLOW.md`).

## Интеграция со складом

При `POST /documents/:id/sign` для типа документа с `warehouse_action` **RESERVE** / **CONSUME** / **RECEIPT** вызывается warehouse по HTTP с заголовками `X-Service-Secret` и `X-Tenant-Id`. В **warehouse-service** должен быть задан тот же секрет (`service_secret`).
