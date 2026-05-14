# Локальный API-шлюз (dev)

Поднимает **nginx** на `http://localhost:8080` и проксирует `/api/v1/...` на микросервисы VKR, запущенные отдельными compose на хосте.

Для большинства бэкендов, если клиент не передаёт заголовок `Authorization`, подставляется `Bearer <cookie access_token>` (как раньше в Caddy).

## Предварительные условия

1. [auth-service](../auth-service): `docker compose up -d` — API слушает **`localhost:18080`** (порт изменён, чтобы освободить **8080** под шлюз). Keycloak: **8081**.
2. [warehouse-service](../warehouse-service), [sed-service](../sed-service), [production-service](../production-service), [procurement-service](../procurement-service), [sales-service](../sales-service), [traceability-service](../traceability-service): каждый свой `docker compose up -d`.

## Запуск

```bash
cd dev-gateway
docker compose up -d
```

Проверка: `curl -s http://localhost:8080/ready` (ответ auth-service).

Фронт [ds-erp-client](../../ds-erp-client): `NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`.

## Проверка конфигурации nginx

Конфиг резолвит `host.docker.internal` при старте контейнера (нужен `extra_hosts` из compose). Проверка синтаксиса:

```bash
cd dev-gateway
docker compose run --rm --no-deps gateway nginx -t
```

## Остановка

```bash
docker compose down
```
