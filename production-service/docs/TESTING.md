# Тестирование production-service

## Юнит-тесты

```bash
make test
```

Пример: расчёт потребности по BOM — `internal/usecases/bom_qty_test.go`.

## E2E

Общий стенд описан в [sed-service/e2e_tests/docker-compose.test.yml](../../sed-service/e2e_tests/docker-compose.test.yml): добавлены `pg-prod`, сервис **production-service** (:28092) и переменные callback для sed.

Запуск из каталога `sed-service/e2e_tests`:

```bash
./run.sh
```

или:

```bash
docker compose -f docker-compose.test.yml up --build -d
pytest -q
```

Smoke-тесты production: `tests/test_production_smoke.py` (`/health`, `/ready`).

Полный сквозной сценарий MES (заказ → release → отчёты → complete) можно расширять отдельным тестом по аналогии с `test_sed_warehouse_flow.py`.
