"""Сценарий продаж/отгрузки: SO→approve (SED callback)→release→reserve→ship→balances."""

import uuid

import requests


def test_sales_so_reserve_ship(auth_api, sed_api, wh_api, sales_api, headers_test, wh_svc_headers):
    # --- auth: superadmin ---
    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    super_tok = r.json()["access_token"]
    h_super = {"Authorization": f"Bearer {super_tok}", "Content-Type": "application/json"}

    tenant = "test_sales_" + uuid.uuid4().hex[:10]
    r = requests.post(
        f"{auth_api}/api/v1/tenants", headers=h_super, json={"code": tenant, "name": "SALES E2E"}, timeout=60
    )
    assert r.status_code == 201, r.text

    # ent-admin
    r = requests.post(
        f"{auth_api}/api/v1/tenants/{tenant}/ent-admin",
        headers=h_super,
        json={"username": "admin", "email": f"admin-{tenant}@test.local", "password": "AdminPass123!"},
        timeout=60,
    )
    assert r.status_code == 201, r.text

    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"admin@{tenant}", "password": "AdminPass123!"},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    ent_tok = r.json()["access_token"]
    h_ent = {"Authorization": f"Bearer {ent_tok}", "Content-Type": "application/json"}

    # sales user
    r = requests.post(
        f"{auth_api}/api/v1/users",
        headers=h_ent,
        json={"username": "manager", "email": f"manager-{tenant}@test.local", "role": "sales_manager"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    uid = r.json()["id"]
    temp_pw = r.json()["temporary_password"]

    roles = ["sales_manager", "sales_viewer", "sed_admin", "sed_author", "sed_approver", "sed_viewer", "warehouse_admin"]
    r = requests.put(
        f"{auth_api}/api/v1/users/{uid}/roles",
        headers=h_ent,
        json={"roles": roles},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"manager@{tenant}", "password": temp_pw},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    user_tok = r.json()["access_token"]
    h_user = {"Authorization": f"Bearer {user_tok}", "Content-Type": "application/json"}

    # --- warehouse seed (service secret) ---
    wh_h = wh_svc_headers(tenant)
    r = requests.post(f"{wh_api}/api/v1/warehouses", headers=wh_h, json={"code": "W1", "name": "Main"}, timeout=60)
    assert r.status_code == 201, r.text
    wh_id = r.json().get("id") or r.json()["ID"]

    r = requests.post(
        f"{wh_api}/api/v1/warehouses/{wh_id}/bins",
        headers=wh_h,
        json={"code": "A1", "name": "Bin", "bin_type": "STORAGE"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    bin_id = r.json().get("id") or r.json()["ID"]

    r = requests.post(
        f"{wh_api}/api/v1/products",
        headers=wh_h,
        json={"sku": "E2E-SALES-1", "name": "SalesProd", "unit": "pcs", "tracking_mode": "NONE"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    product_id = r.json().get("id") or r.json()["ID"]

    # stock receipt so that reserve+ship is possible
    r = requests.post(
        f"{wh_api}/api/v1/operations/receipt",
        headers=wh_h,
        json={"warehouse_id": wh_id, "bin_id": bin_id, "lines": [{"product_id": product_id, "qty": "10", "unit_cost": "1"}]},
        timeout=60,
    )
    assert r.status_code == 201, r.text

    # --- SED: workflow + doc type SALES_ORDER_APPROVAL (warehouse_action NONE) ---
    r = requests.post(f"{sed_api}/api/v1/workflows", headers=h_user, json={"code": "WF-SALES", "name": "Sales flow"}, timeout=60)
    assert r.status_code == 201, r.text
    wf_id = r.json()["id"]

    r = requests.post(
        f"{sed_api}/api/v1/workflows/{wf_id}/steps",
        headers=h_user,
        json={"order_no": 1, "name": "Approve", "required_role": "sed_approver"},
        timeout=60,
    )
    assert r.status_code == 201, r.text

    r = requests.post(
        f"{sed_api}/api/v1/document-types",
        headers=h_user,
        json={"code": "SALES_ORDER_APPROVAL", "name": "SO approval", "warehouse_action": "NONE", "default_workflow_id": wf_id},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    so_type_id = r.json()["id"]

    # --- sales: customer + SO + line ---
    r = requests.post(
        f"{sales_api}/api/v1/customers",
        headers=h_user,
        json={"code": "C1", "name": "Customer 1", "active": True, "contacts": {}},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    customer_id = r.json()["id"]

    r = requests.post(
        f"{sales_api}/api/v1/sales-orders",
        headers=h_user,
        json={"customer_id": customer_id, "ship_from_warehouse_id": wh_id, "ship_from_bin_id": bin_id},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    so_id = r.json()["id"]

    r = requests.post(
        f"{sales_api}/api/v1/sales-orders/{so_id}/lines",
        headers=h_user,
        json={"line_no": 1, "product_id": product_id, "qty": "3", "uom": "pcs"},
        timeout=60,
    )
    assert r.status_code == 201, r.text

    # submit SO -> approve+sign in SED -> callback -> SO APPROVED
    r = requests.post(
        f"{sales_api}/api/v1/sales-orders/{so_id}/submit",
        headers=h_user,
        json={"sed_document_type_id": so_type_id, "title": "SO approve"},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    r = requests.get(f"{sales_api}/api/v1/sales-orders/{so_id}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    sed_doc_id = r.json()["so"]["sed_document_id"]

    r = requests.post(f"{sed_api}/api/v1/documents/{sed_doc_id}/approve", headers=h_user, json={"comment": "ok"}, timeout=60)
    assert r.status_code == 204, r.text
    r = requests.post(f"{sed_api}/api/v1/documents/{sed_doc_id}/sign", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    r = requests.get(f"{sales_api}/api/v1/sales-orders/{so_id}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    assert r.json()["so"]["status"] == "APPROVED"

    r = requests.post(f"{sales_api}/api/v1/sales-orders/{so_id}/release", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    r = requests.post(f"{sales_api}/api/v1/sales-orders/{so_id}/reserve", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    assert r.json()["reservation_ids"]

    r = requests.post(f"{sales_api}/api/v1/sales-orders/{so_id}/ship", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    assert r.json()["warehouse_document_id"]

    # balances should be reduced (qty 10 -> 7)
    r = requests.get(f"{wh_api}/api/v1/balances?product_id={product_id}&warehouse_id={wh_id}&bin_id={bin_id}", headers=wh_h, timeout=60)
    assert r.status_code == 200, r.text
    bals = r.json() or []

    def _qty(row):
        return row.get("qty") or row.get("quantity") or row.get("Quantity")

    # at least one row has qty <= 7 (depends on batch handling)
    assert any(float(str(_qty(b))) <= 7.0 for b in bals)

