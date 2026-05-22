# RAG — руководство для команды

Корпус из **50 документов СЭД** (10 типов × 5), **11 учётных записей** (включая сервисный аккаунт) с разным ACL и JSON-manifest для индексации и проверки фильтрации по ролям.

**Формат логина:** `{username}@ragcorp`  
**Общий пароль всех** `rag_`***:**  Уточнить у администратора

**Базовый URL стенда (один порт для всего API и Keycloak):**

```text
http://85.236.191.21:52556/
```

---

## Учётные записи


| Username           | Логин                      | Пароль                    | Роли                                  | `GET /api/v1/documents`                                    |
| ------------------ | -------------------------- | ------------------------- | ------------------------------------- | ---------------------------------------------------------- |
| `**rag_service**`  | `rag_service@ragcorp`      | Уточнить у администратора | `sed_admin`                           | **200**, **50** — **сервисный аккаунт RAG**, полный доступ |
| `rag_admin`        | `rag_admin@ragcorp`        | тот же                    | `sed_admin`, `sed_author`             | **200**, **50**                                            |
| `rag_finance`      | `rag_finance@ragcorp`      | тот же                    | `doc_read_finance`                    | **200**, **15** (PR, PO, SC)                               |
| `rag_procurement`  | `rag_procurement@ragcorp`  | тот же                    | `doc_read_procurement`                | **200**, **15**                                            |
| `rag_sales`        | `rag_sales@ragcorp`        | тот же                    | `doc_read_sales`                      | **200**, **10** (SO, SH)                                   |
| `rag_production`   | `rag_production@ragcorp`   | тот же                    | `doc_read_production`                 | **200**, **10**                                            |
| `rag_warehouse`    | `rag_warehouse@ragcorp`    | тот же                    | `doc_read_warehouse`                  | **200**, **15**                                            |
| `rag_author_proc`  | `rag_author_proc@ragcorp`  | тот же                    | `sed_author`, `doc_write_procurement` | см. сценарии ниже                                          |
| `rag_author_sales` | `rag_author_sales@ragcorp` | тот же                    | `sed_author`, `doc_write_sales`       | см. сценарии ниже                                          |
| `rag_approver`     | `rag_approver@ragcorp`     | тот же                    | `sed_approver`                        | **200**, **0**; `GET /tasks`                               |
| `rag_no_access`    | `rag_no_access@ragcorp`    | тот же                    | `sed_viewer`                          | **403**                                                    |


### Сервисный аккаунт `rag_service`

Для индексации и batch-запросов RAG-модуля: роль `**sed_admin`** даёт чтение **всех 50** документов тенанта без фильтра `doc_read_`*. 

### Тестовая ручка для обхода авторизации.

```bash
curl -s -X POST "$API_BASE/api/v1/internal/test/login" \
  -H "X-Test-Secret: SECRET" -H "Content-Type: application/json" \
  -d '{"username":"rag_service@ragcorp","password":"PASSWORD"}'
```

Secret & Password уточнить у администратора

### Какие типы документов видит роль


| Код типа                     | Префикс `rag_id` | Кто читает (`doc_read_*`) |
| ---------------------------- | ---------------- | ------------------------- |
| `PURCHASE_REQUEST_APPROVAL`  | `RAG-PR-xxx`     | procurement, finance      |
| `PURCHASE_ORDER_APPROVAL`    | `RAG-PO-xxx`     | procurement, finance      |
| `SUPPLIER_CONTRACT_APPROVAL` | `RAG-SC-xxx`     | procurement, finance      |
| `SALES_ORDER_APPROVAL`       | `RAG-SO-xxx`     | sales                     |
| `SHIPMENT_APPROVAL`          | `RAG-SH-xxx`     | sales                     |
| `BOM_APPROVAL`               | `RAG-BM-xxx`     | production                |
| `ROUTING_APPROVAL`           | `RAG-RT-xxx`     | production                |
| `RAG_WH_RESERVE`             | `RAG-WR-xxx`     | warehouse                 |
| `RAG_WH_CONSUME`             | `RAG-WC-xxx`     | warehouse                 |
| `RAG_WH_RECEIPT`             | `RAG-WT-xxx`     | warehouse                 |


---

### Индексация

Публичный `GET /api/v1/documents` отдаёт **только бизнес-карточку** (`title`, `payload` с заказами, поставщиками, строками и т.д.)

Корпус для RAG и матрица видимости — **закрытая сервисная ручка**:

```http
GET http://<хост>:52556/api/v1/internal/rag/corpus
X-Service-Secret: e2e-service-secret
X-Tenant-Id: ragcorp
```

Отдает json 

```
{
  "documents": [
    {
      "document_id": "a4f1f756-4eea-41a7-86ec-95444e446c27",
      "text": "Заявка на закупку RAG-PR-001 — Подшипник 6205-2RS\n\nПодразделение: Отдел снабжения\nИнициатор: Иванов А.С.\nОбоснование: Плановое пополнение...\nПозиции:\n  1) BRG-6205-2RS Подшипник — 90 шт ...",
      "access": {
        "read_roles": ["doc_read_procurement", "doc_read_finance"],
        "write_roles": ["doc_write_procurement"],
        "approve_roles": ["sed_approver"],
        "admin_roles": ["sed_admin"]
      },
      "attachments": [
        {
          "file_id": "...",
          "name": "договор.pdf",
          "content_type": "application/pdf",
          "size_bytes": 12345,
          "url": "http://85.236.191.21:52556/api/v1/internal/rag/documents/{doc_id}/files/{file_id}"
        }
      ]
    }
  ]
}
```

Где:

- text полное содеражние документа
- access список ролей для чтения, записи, апрува, администрирования
- attachments информация о файле и ссылка для его скачивания

