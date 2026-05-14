# Локальный API-шлюз (dev)

Поднимает **Caddy** на `http://localhost:8080` и проксирует `/api/v1/...` на микросервисы VKR, запущенные отдельными compose на хосте.

## Предварительные условия

1. [auth-service](../auth-service): `docker compose up -d` — API слушает **`localhost:18080`** (порт изменён, чтобы освободить **8080** под шлюз). Keycloak: **8081**.
2. [warehouse-service](../warehouse-service), [sed-service](../sed-service), [production-service](../production-service), [procurement-service](../procurement-service): каждый свой `docker compose up -d`.

## Запуск

```bash
cd dev-gateway
docker compose up -d
```

Проверка: `curl -s http://localhost:8080/ready` (ответ auth-service).

Фронт [ds-erp-client](../../ds-erp-client): `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`.

## Остановка

```bash
docker compose down
```
