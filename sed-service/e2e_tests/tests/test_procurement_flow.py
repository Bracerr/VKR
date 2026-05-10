"""Сценарий закупок: PR→PO→receive→balances + callback SIGNED от SED."""

import uuid

import requests


def test_procurement_pr_po_receive(auth_api, sed_api, wh_api, proc_api, headers_test, wh_svc_headers):
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

    tenant = "test_proc_" + uuid.uuid4().hex[:10]
    r = requests.post(
        f"{auth_api}/api/v1/tenants", headers=h_super, json={"code": tenant, "name": "PROC E2E"}, timeout=60
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

    # procurement user
    r = requests.post(
        f"{auth_api}/api/v1/users",
        headers=h_ent,
        json={"username": "buyer", "email": f"buyer-{tenant}@test.local", "role": "proc_buyer"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    uid = r.json()["id"]
    temp_pw = r.json()["temporary_password"]

    roles = ["proc_buyer", "proc_viewer", "sed_admin", "sed_author", "sed_approver", "sed_viewer", "warehouse_admin"]
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
        json={"username": f"buyer@{tenant}", "password": temp_pw},
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
        json={"sku": "E2E-PROC-1", "name": "ProcProd", "unit": "pcs", "tracking_mode": "NONE"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    product_id = r.json().get("id") or r.json()["ID"]

    # --- SED: workflow + doc types for procurement (warehouse_action NONE) ---
    r = requests.post(f"{sed_api}/api/v1/workflows", headers=h_user, json={"code": "WF-PROC", "name": "Proc flow"}, timeout=60)
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
        json={"code": "PURCHASE_REQUEST_APPROVAL", "name": "PR approval", "warehouse_action": "NONE", "default_workflow_id": wf_id},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    pr_type_id = r.json()["id"]

    r = requests.post(
        f"{sed_api}/api/v1/document-types",
        headers=h_user,
        json={"code": "PURCHASE_ORDER_APPROVAL", "name": "PO approval", "warehouse_action": "NONE", "default_workflow_id": wf_id},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    po_type_id = r.json()["id"]

    # --- procurement: supplier ---
    r = requests.post(
        f"{proc_api}/api/v1/suppliers",
        headers=h_user,
        json={"code": "S1", "name": "Supplier 1", "active": True, "contacts": {}},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    supplier_id = r.json()["id"]

    # --- procurement: PR + line + submit ---
    r = requests.post(f"{proc_api}/api/v1/purchase-requests", headers=h_user, json={}, timeout=60)
    assert r.status_code == 201, r.text
    pr_id = r.json()["id"]

    r = requests.post(
        f"{proc_api}/api/v1/purchase-requests/{pr_id}/lines",
        headers=h_user,
        json={"line_no": 1, "product_id": product_id, "qty": "3", "uom": "pcs", "target_warehouse_id": wh_id, "target_bin_id": bin_id},
        timeout=60,
    )
    assert r.status_code == 201, r.text

    r = requests.post(
        f"{proc_api}/api/v1/purchase-requests/{pr_id}/submit",
        headers=h_user,
        json={"sed_document_type_id": pr_type_id, "title": "PR approve"},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    # approve+sign in sed (triggers callback → PR APPROVED)
    r = requests.get(f"{proc_api}/api/v1/purchase-requests/{pr_id}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    sed_doc_id = r.json()["pr"]["sed_document_id"]

    r = requests.post(f"{sed_api}/api/v1/documents/{sed_doc_id}/approve", headers=h_user, json={"comment": "ok"}, timeout=60)
    assert r.status_code == 204, r.text
    r = requests.post(f"{sed_api}/api/v1/documents/{sed_doc_id}/sign", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    r = requests.get(f"{proc_api}/api/v1/purchase-requests/{pr_id}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    assert r.json()["pr"]["status"] == "APPROVED"

    # --- procurement: PO from PR + submit ---
    r = requests.post(
        f"{proc_api}/api/v1/purchase-orders/from-pr/{pr_id}",
        headers=h_user,
        json={"supplier_id": supplier_id},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    po_id = r.json()["id"]

    r = requests.post(
        f"{proc_api}/api/v1/purchase-orders/{po_id}/submit",
        headers=h_user,
        json={"sed_document_type_id": po_type_id, "title": "PO approve"},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    r = requests.get(f"{proc_api}/api/v1/purchase-orders/{po_id}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    sed_po_doc_id = r.json()["po"]["sed_document_id"]

    r = requests.post(f"{sed_api}/api/v1/documents/{sed_po_doc_id}/approve", headers=h_user, json={"comment": "ok"}, timeout=60)
    assert r.status_code == 204, r.text
    r = requests.post(f"{sed_api}/api/v1/documents/{sed_po_doc_id}/sign", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    # release PO after callback approval
    r = requests.post(f"{proc_api}/api/v1/purchase-orders/{po_id}/release", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    # --- receive (warehouse receipt) ---
    r = requests.post(
        f"{proc_api}/api/v1/purchase-orders/{po_id}/receive",
        headers=h_user,
        json={"warehouse_id": wh_id, "bin_id": bin_id},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    wh_doc_id = r.json()["warehouse_document_id"]
    assert wh_doc_id

    # balances should reflect receipt
    r = requests.get(f"{wh_api}/api/v1/balances", headers=wh_h, timeout=60)
    assert r.status_code == 200, r.text
    balances = r.json() or []
    def _pid(row):
        return row.get("product_id") or row.get("ProductID")

    def _qty(row):
        return row.get("qty") or row.get("quantity") or row.get("Quantity")

    assert any(_pid(b) == product_id and str(_qty(b)) not in ("0", "0.0", "0.00000000", "None", "null", "") for b in balances)

