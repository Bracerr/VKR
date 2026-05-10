# Тестирование

## Юнит-тесты

```bash
make test
# или
go test ./... -race -count=1
```

Пакет `internal/usecases` содержит тесты сценария склада на подписи (`RunWarehouseOnSign`) и вспомогательной логики.

## E2E (pytest)

Из корня репозитория **VKR** (родитель каталогов `auth-service`, `warehouse-service`, `sed-service`):

```bash
docker compose -f sed-service/e2e_tests/docker-compose.test.yml up --build -d
# дождаться готовности Keycloak и сервисов (несколько минут)
python3 -m venv .venv && . .venv/bin/activate
pip install -r sed-service/e2e_tests/requirements.txt
pytest -q sed-service/e2e_tests/tests
docker compose -f sed-service/e2e_tests/docker-compose.test.yml down -v
```

Переменные окружения (по умолчанию совпадают с портами compose):

- `AUTH_BASE_URL` — `http://localhost:28080`
- `WAREHOUSE_BASE_URL` — `http://localhost:28090`
- `SED_BASE_URL` — `http://localhost:28091`
- `TEST_SECRET` — `e2e-test-secret` (как в auth-service)
- `WAREHOUSE_SERVICE_SECRET` — `sed-e2e-wh-secret` (совпадает с `SERVICE_SECRET` warehouse в compose)

Сценарий: bootstrap тенанта и пользователя через auth, приход на склад, создание маршрута и типа документа RESERVE в СЭД, submit → approve → sign, проверка появления резерва в warehouse.

## Swagger

```bash
make swagger
```
