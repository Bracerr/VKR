# sales-service

Контур **продаж и отгрузки**: заказ клиента (SO) → согласование в `sed-service` → резерв в `warehouse-service` → отгрузка (списание со склада).

## Быстрый старт (локально)

- PostgreSQL на `5437`
- HTTP на `8094`

```bash
cd sales-service
cp .env.example .env
make up
```

Проверка:

```bash
curl -s http://localhost:8094/health
curl -s http://localhost:8094/ready
```

## Документация

- [`docs/SALES.md`](docs/SALES.md)
- [`docs/TESTING.md`](docs/TESTING.md)

