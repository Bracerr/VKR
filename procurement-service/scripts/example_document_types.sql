-- Пример seed для SED: типы документов закупок.
-- Выполняется в БД SED в рамках нужного tenant_code.
-- Привязку к маршрутам согласования (workflows) делайте через UI/API SED.

-- Замените :tenant на нужный tenant_code (например 'demo').
-- Сначала создайте workflows и получите их id, затем подставьте в default_workflow_id.

-- PURCHASE_REQUEST_APPROVAL
INSERT INTO document_types(id, tenant_code, code, name, warehouse_action, default_workflow_id)
VALUES (gen_random_uuid(), :tenant, 'PURCHASE_REQUEST_APPROVAL', 'Согласование заявки на закупку (PR)', 'NONE', NULL);

-- PURCHASE_ORDER_APPROVAL
INSERT INTO document_types(id, tenant_code, code, name, warehouse_action, default_workflow_id)
VALUES (gen_random_uuid(), :tenant, 'PURCHASE_ORDER_APPROVAL', 'Согласование заказа поставщику (PO)', 'NONE', NULL);

-- SUPPLIER_CONTRACT_APPROVAL (опционально)
INSERT INTO document_types(id, tenant_code, code, name, warehouse_action, default_workflow_id)
VALUES (gen_random_uuid(), :tenant, 'SUPPLIER_CONTRACT_APPROVAL', 'Согласование договора с поставщиком', 'NONE', NULL);

