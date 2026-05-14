"""Traceability: warehouse receipt -> reservation -> issue -> trace graph."""

import time
import uuid

import requests


def test_traceability_warehouse_flow(auth_api, wh_api, trace_api, headers_test, wh_svc_headers):
    # superadmin login
    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    super_tok = r.json()["access_token"]
    h_super = {"Authorization": f"Bearer {super_tok}", "Content-Type": "application/json"}

    tenant = "test_trace_" + uuid.uuid4().hex[:10]
    r = requests.post(
        f"{auth_api}/api/v1/tenants", headers=h_super, json={"code": tenant, "name": "TRACE E2E"}, timeout=60
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

    # create user with warehouse_viewer (enough for trace read) + warehouse_admin to read warehouse if needed
    r = requests.post(
        f"{auth_api}/api/v1/users",
        headers=h_ent,
        json={"username": "viewer", "email": f"viewer-{tenant}@test.local", "role": "warehouse_viewer"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    uid = r.json()["id"]
    temp_pw = r.json()["temporary_password"]

    r = requests.put(
        f"{auth_api}/api/v1/users/{uid}/roles",
        headers=h_ent,
        json={"roles": ["warehouse_viewer"]},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"viewer@{tenant}", "password": temp_pw},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    user_tok = r.json()["access_token"]
    h_user = {"Authorization": f"Bearer {user_tok}", "Content-Type": "application/json"}

    # seed warehouse (service secret)
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

    # SERIAL product
    r = requests.post(
        f"{wh_api}/api/v1/products",
        headers=wh_h,
        json={"sku": "E2E-TRACE-1", "name": "TraceProd", "unit": "pcs", "tracking_mode": "SERIAL"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    product_id = r.json().get("id") or r.json()["ID"]

    sn1 = "SN-" + uuid.uuid4().hex[:8]
    r = requests.post(
        f"{wh_api}/api/v1/operations/receipt",
        headers=wh_h,
        json={
            "warehouse_id": wh_id,
            "bin_id": bin_id,
            "lines": [{"product_id": product_id, "qty": "1", "unit_cost": "1", "serial_numbers": [sn1]}],
        },
        timeout=60,
    )
    assert r.status_code == 201, r.text

    # reserve by serial_no
    r = requests.post(
        f"{wh_api}/api/v1/reservations",
        headers=wh_h,
        json={
            "warehouse_id": wh_id,
            "bin_id": bin_id,
            "product_id": product_id,
            "qty": "1",
            "serial_no": sn1,
            "reason": "e2e",
            "doc_ref": "trace",
        },
        timeout=60,
    )
    assert r.status_code == 201, r.text
    res_id = r.json()["id"]

    # issue from reservations -> should trigger trace doc posted event
    r = requests.post(
        f"{wh_api}/api/v1/operations/issue-from-reservations",
        headers=wh_h,
        json={"reservation_ids": [res_id]},
        timeout=60,
    )
    assert r.status_code == 201, r.text

    # allow async callbacks a moment
    time.sleep(0.5)

    # search by serial_no
    r = requests.get(f"{trace_api}/api/v1/trace/search?serial_no={sn1}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    anchors = r.json().get("anchors") or []
    assert anchors, r.text

