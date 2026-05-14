# traceability-service

Сервис-агрегатор для сквозной прослеживаемости (Traceability) по событиям из доменных сервисов и `warehouse-service`.

## Запуск локально

```bash
docker compose up -d --build
```

Проверка:

- `GET /health`
- `GET /ready`

## API (MVP)

Публичные (JWT):

- `GET /api/v1/trace/search?serial_no=&batch_id=&product_id=&from=&to=`
- `GET /api/v1/trace/graph?anchor_type=&anchor_id=&from=&to=&depth=`

Внутренние (service-secret):

- `POST /api/v1/internal/events`

## Документация

См. `docs/TRACE.md`.

