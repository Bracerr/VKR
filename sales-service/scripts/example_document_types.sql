-- Пример seed для SED: типы документов продаж/отгрузки.
-- Выполняется в БД SED в рамках нужного tenant_code.
-- Привязку к маршрутам согласования (workflows) делайте через UI/API SED.

-- Замените :tenant на нужный tenant_code (например 'demo').

-- SALES_ORDER_APPROVAL
INSERT INTO document_types(id, tenant_code, code, name, warehouse_action, default_workflow_id)
VALUES (gen_random_uuid(), :tenant, 'SALES_ORDER_APPROVAL', 'Согласование заказа клиента (SO)', 'NONE', NULL);

-- SHIPMENT_APPROVAL (опционально)
INSERT INTO document_types(id, tenant_code, code, name, warehouse_action, default_workflow_id)
VALUES (gen_random_uuid(), :tenant, 'SHIPMENT_APPROVAL', 'Согласование отгрузки', 'NONE', NULL);

