# Тестирование sales-service

## Юнит-тесты (Go)

Требуется локальный Postgres на `localhost:5437` (как в `docker-compose.yaml`).

```bash
cd sales-service
make up
go test ./... -count=1
```

## E2E (pytest)

E2E запускаются из `sed-service/e2e_tests` (общий стенд всех сервисов).

- compose-файл: `sed-service/e2e_tests/docker-compose.test.yml`
- тесты: `sed-service/e2e_tests/tests/test_sales_smoke.py`, `sed-service/e2e_tests/tests/test_sales_flow.py`

Пример:

```bash
cd sed-service/e2e_tests
. .venv/bin/activate
docker compose -f docker-compose.test.yml up -d --build
pytest -q -k sales_
```

