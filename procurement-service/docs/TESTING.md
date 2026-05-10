# Тестирование procurement-service

## Юнит

```bash
cd procurement-service
make test
```

Примечание: тесты usecases используют локальный Postgres `localhost:5436` и при его отсутствии будут пропущены.

## E2E (pytest)

E2E сценарий закупок живёт в `sed-service/e2e_tests` и поднимает полный стенд (auth + keycloak + warehouse + sed + production + procurement).

```bash
cd sed-service/e2e_tests
docker compose -f docker-compose.test.yml up -d --build

python3 -m venv .venv
. .venv/bin/activate
pip install -r requirements.txt
pytest -q

docker compose -f docker-compose.test.yml down -v
```

