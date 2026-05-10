"""Создание заявки СЭД → согласование → подпись → резерв на складе."""
import uuid

import requests


def test_reserve_on_sign(auth_api, sed_api, wh_api, headers_test, wh_svc_headers):
    # --- auth: tenant + ent-admin + пользователь с ролями СЭД и складом ---
    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    super_tok = r.json()["access_token"]
    h_super = {"Authorization": f"Bearer {super_tok}", "Content-Type": "application/json"}

    tenant = "test_sed_" + uuid.uuid4().hex[:10]
    r = requests.post(
        f"{auth_api}/api/v1/tenants", headers=h_super, json={"code": tenant, "name": "SED E2E"}, timeout=60
    )
    assert r.status_code == 201, r.text

    r = requests.post(
        f"{auth_api}/api/v1/tenants/{tenant}/ent-admin",
        headers=h_super,
        json={
            "username": "admin",
            "email": f"admin-{tenant}@test.local",
            "password": "AdminPass123!",
        },
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

    r = requests.post(
        f"{auth_api}/api/v1/users",
        headers=h_ent,
        json={
            "username": "seduser",
            "email": f"seduser-{tenant}@test.local",
            "role": "sed_admin",
        },
        timeout=60,
    )
    assert r.status_code == 201, r.text
    uid = r.json()["id"]
    temp_pw = r.json()["temporary_password"]

    roles = [
        "sed_admin",
        "sed_author",
        "sed_approver",
        "sed_viewer",
        "warehouse_admin",
    ]
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
        json={"username": f"seduser@{tenant}", "password": temp_pw},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    user_tok = r.json()["access_token"]
    h_user = {"Authorization": f"Bearer {user_tok}", "Content-Type": "application/json"}

    # --- склад: справочники и приход (service secret) ---
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
        json={"sku": "E2E-1", "name": "Prod", "unit": "pcs", "tracking_mode": "NONE"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    product_id = r.json().get("id") or r.json()["ID"]

    r = requests.post(
        f"{wh_api}/api/v1/operations/receipt",
        headers=wh_h,
        json={
            "warehouse_id": wh_id,
            "bin_id": bin_id,
            "lines": [{"product_id": product_id, "qty": "100", "unit_cost": "1"}],
        },
        timeout=60,
    )
    assert r.status_code == 201, r.text

    r = requests.get(f"{wh_api}/api/v1/reservations", headers=wh_h, timeout=60)
    assert r.status_code == 200, r.text
    n_res_before = len(r.json() or [])

    # --- СЭД: маршрут, тип, документ ---
    r = requests.post(f"{sed_api}/api/v1/workflows", headers=h_user, json={"code": "WF1", "name": "One step"}, timeout=60)
    assert r.status_code == 201, r.text
    wf_id = r.json()["id"]

    r = requests.post(
        f"{sed_api}/api/v1/workflows/{wf_id}/steps",
        headers=h_user,
        json={
            "order_no": 1,
            "name": "Approve",
            "required_role": "sed_approver",
        },
        timeout=60,
    )
    assert r.status_code == 201, r.text

    r = requests.post(
        f"{sed_api}/api/v1/document-types",
        headers=h_user,
        json={
            "code": "ZAYAVKA",
            "name": "Заявка",
            "warehouse_action": "RESERVE",
            "default_workflow_id": wf_id,
        },
        timeout=60,
    )
    assert r.status_code == 201, r.text
    type_id = r.json()["id"]

    payload = {
        "warehouse_id": wh_id,
        "default_bin_id": bin_id,
        "lines": [
            {
                "product_id": product_id,
                "qty": "3",
                "reason": "e2e",
                "doc_ref": "sed-e2e",
            }
        ],
    }
    r = requests.post(
        f"{sed_api}/api/v1/documents",
        headers=h_user,
        json={"type_id": type_id, "title": "Отгрузка", "payload": payload},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    doc_id = r.json()["id"]

    r = requests.post(f"{sed_api}/api/v1/documents/{doc_id}/submit", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    r = requests.post(
        f"{sed_api}/api/v1/documents/{doc_id}/approve",
        headers=h_user,
        json={"comment": "ok"},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    r = requests.post(f"{sed_api}/api/v1/documents/{doc_id}/sign", headers=h_user, timeout=60)
    assert r.status_code == 204, r.text

    r = requests.get(f"{sed_api}/api/v1/documents/{doc_id}", headers=h_user, timeout=60)
    assert r.status_code == 200, r.text
    body = r.json()
    assert body["status"] == "SIGNED"
    assert "reservation_ids" in body["warehouse_ref"]

    r = requests.get(f"{wh_api}/api/v1/reservations", headers=wh_h, timeout=60)
    assert r.status_code == 200, r.text
    assert len(r.json() or []) > n_res_before
