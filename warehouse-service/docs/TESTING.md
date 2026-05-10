# Тестирование warehouse-service

## Юнит-тесты

```bash
make test
# или
go test ./... -count=1
```

- `internal/usecases/reports_uc_test.go` — классификация ABC.
- Интеграция (приёмка + FEFO + нехватка): задаётся DSN **поднятой** БД с применёнными миграциями:

```bash
export WAREHOUSE_TEST_DSN='postgres://wh:wh@localhost:5433/warehouse?sslmode=disable'
go test ./internal/usecases -count=1 -run Integration
```

## Ручная проверка (curl)

1. Получить JWT (тот же Keycloak, в токене должны быть `tenant_id` и одна из ролей склада).
2. Пример:

```bash
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8090/api/v1/products
```

## E2E (pytest)

```bash
cd e2e_tests
python3 -m venv .venv && . .venv/bin/activate
pip install -r requirements.txt
export WAREHOUSE_URL=http://localhost:8090
pytest -q
```

По умолчанию тесты только **smoke** (`/health`). Полный сценарий с JWT и auth-service можно расширить при наличии поднятого стенда (см. `docker-compose.test.yml` как заготовку).

## Линтер

```bash
make lint
```
