# auth-service

Микросервис авторизации и управления пользователями для системы документооборота (мультитенантность через **Keycloak**).

## Стек

- Go 1.22+, Gin, чистая архитектура (`handlers` / `usecases` / `repositories` / `models`)
- PostgreSQL, golang-migrate, Keycloak 24 (realm `industrial-sed`)
- OIDC **Authorization Code + PKCE** (BFF, httpOnly cookies), внутренний **password grant** только для `/internal/test/login` (e2e)
- JWT-проверка по JWKS (`jwtverify` + keyfunc), Admin API — `gocloak`
- Swagger: `make swagger` → [docs/swagger.json](docs/swagger.json), UI: `/swagger/index.html`

## Быстрый старт (Docker)

```bash
cp .env.example .env   # при необходимости поправьте секреты
make up               # postgres + keycloak + auth-service
```

HTTP API по умолчанию проброшен на хост**:18080** (порт **8080** на хосте зарезервирован под единый шлюз **nginx** — см. [`dev-gateway/`](../dev-gateway/README.md) и `docker-compose.yaml`).

Подождите ~60 с и проверьте:

```bash
curl -s http://localhost:18080/ready
```

Swagger: <http://localhost:18080/swagger/index.html>

По умолчанию создаётся суперадмин (если `BOOTSTRAP_SUPERADMIN=true` в compose):

- логин: `superadmin`
- пароль: `superadmin`

## Локальный запуск без Docker (только бинарник)

Нужны PostgreSQL, Keycloak и:

- YAML-конфиг (обычные настройки): `configs/local.yaml`
- `.env` (секреты): см. `.env.example`

```bash
export CONFIG_PATH=./configs/local.yaml
make migrate-up
make run
```

## Makefile

| Цель | Описание |
|------|----------|
| `make run` | `go run ./cmd/auth-service` |
| `make build` | сборка в `bin/auth-service` |
| `make test` | юнит-тесты |
| `make test-cover` | покрытие |
| `make lint` | golangci-lint |
| `make swagger` | генерация Swagger |
| `make mocks` | gomock для `ports.KeycloakClient` и репозиториев |
| `make migrate-up` / `migrate-down` | миграции (`DB_DSN`) |
| `make up` / `make down` | docker compose |
| `make e2e` | pytest + отдельный compose (`e2e_tests/`) |

## Конфигурация (YAML) и секреты (.env)

Обычные настройки лежат в YAML и выбираются через `CONFIG_PATH`:

- Docker: `configs/docker.yaml`
- Локально: `configs/local.yaml`

Секреты держим в `.env`:

- `DB_DSN`
- `KEYCLOAK_ADMIN_PASSWORD`
- `KEYCLOAK_CLIENT_SECRET`
- `SERVICE_SECRET`
- `TEST_SECRET`
- `STATE_COOKIE_SECRET`
- `BOOTSTRAP_SUPERADMIN_PASSWORD`

## Bootstrap первого `ent_admin` предприятия

После `POST /api/v1/tenants` суперадмин может вызвать:

`POST /api/v1/tenants/{code}/ent-admin` с телом:

```json
{"username":"admin","email":"admin@corp.ru","password":"НадёжныйПароль123"}
```

Создаётся пользователь `admin@{code}` с ролью `ent_admin` в группе `tenant_{code}`.

## Документация

- [docs/FLOW.md](docs/FLOW.md) — полный флоу: супер-админ, тенанты, ent_admin, пользователи, роли
- [docs/TESTING.md](docs/TESTING.md) — как проверять API вручную и через pytest
- [docs/FRONTEND.md](docs/FRONTEND.md) — интеграция SPA (BFF cookies, axios, ошибки)

## CI

GitHub Actions: [.github/workflows/ci.yml](.github/workflows/ci.yml) — `golangci-lint` и `go test`.

Если репозиторий — монорепозиторий, задайте `working-directory: auth-service` в шагах workflow.
