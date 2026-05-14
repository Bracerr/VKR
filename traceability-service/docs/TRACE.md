# Traceability (MVP)

## Зачем нужен сервис

`traceability-service` собирает события из:

- `warehouse-service` — **DocumentPosted**: проведённые складские документы и движения (с `batch_id`/`serial_id`)
- доменных сервисов (`sales/procurement/production`) — **LinkEntityWarehouseDoc**: связь “наш документ → warehouse_document_id”

На выходе — быстрый API для UI: поиск якоря и построение графа связей.

## События ingest

### 1) DocumentPosted (из `warehouse-service`)

`POST /api/v1/internal/events`

```json
{
  "event_type": "DocumentPosted",
  "tenant_code": "tenant1",
  "idempotency_key": "wh-doc-<uuid>",
  "payload": {
    "document_id": "<uuid>",
    "doc_type": "RECEIPT|ISSUE|TRANSFER|INVENTORY|...",
    "number": "optional",
    "posted_at": "2026-01-01T00:00:00Z",
    "lines": [
      {
        "product_id": "<uuid>",
        "batch_id": "<uuid>",
        "batch_series": "optional",
        "serial_id": "<uuid>",
        "serial_no": "optional",
        "qty": "1.000"
      }
    ]
  }
}
```

### 2) LinkEntityWarehouseDoc (из доменных сервисов)

```json
{
  "event_type": "LinkEntityWarehouseDoc",
  "tenant_code": "tenant1",
  "idempotency_key": "so-link-<uuid>",
  "payload": {
    "entity_type": "SO|PO|PROD_ORDER",
    "entity_id": "<uuid>",
    "entity_number": "optional",
    "warehouse_document_id": "<uuid>"
  }
}
```

## Граф (как храним)

Таблицы:

- `trace_events` — сырые события (аудит/идемпотентность)
- `trace_nodes` — нормализованные узлы (`WAREHOUSE_DOC`, `BATCH`, `SERIAL`, `SO`, `PO`, `PROD_ORDER`, ...)
- `trace_edges` — связи (`DOC_HAS_SERIAL`, `DOC_HAS_BATCH`, `ENTITY_POSTED_AS_DOC`, ...)

## Публичные API

### Search

`GET /api/v1/trace/search?serial_no=&batch_id=&product_id=&from=&to=`

MVP: возвращает список якорных узлов:

- по `batch_id` — узел `BATCH` с `external_id = batch_id`
- по `serial_no` — узлы `SERIAL` где `meta.serial_no == serial_no`

### Graph

`GET /api/v1/trace/graph?anchor_type=&anchor_id=&from=&to=&depth=`

MVP: BFS по `trace_edges` до указанной глубины (1..6). Фильтры `from/to` пока не режут граф (задел под следующий шаг).

