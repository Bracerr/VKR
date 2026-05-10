"""Сценарий 2: тенант → ent_admin → пользователь → логин → userinfo."""
import requests


def test_login_and_userinfo(api, headers_test, headers_service):
    # super
    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    super_tok = r.json()["access_token"]
    h_super = {"Authorization": f"Bearer {super_tok}", "Content-Type": "application/json"}

    tenant = "test_login_rom"
    requests.delete(f"{api}/api/v1/tenants/{tenant}", headers=h_super, timeout=30)  # ignore

    r = requests.post(
        f"{api}/api/v1/tenants", headers=h_super, json={"code": tenant, "name": "Rom"}, timeout=30
    )
    assert r.status_code in (201, 409), r.text

    r = requests.post(
        f"{api}/api/v1/tenants/{tenant}/ent-admin",
        headers=h_super,
        json={"username": "admin", "email": "admin@test.local", "password": "AdminPass123!"},
        timeout=30,
    )
    assert r.status_code == 201, r.text

    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"admin@{tenant}", "password": "AdminPass123!"},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    ent_tok = r.json()["access_token"]
    h_ent = {"Authorization": f"Bearer {ent_tok}", "Content-Type": "application/json"}

    r = requests.post(
        f"{api}/api/v1/users",
        headers=h_ent,
        json={"username": "ivan", "email": "ivan@test.local", "role": "engineer"},
        timeout=30,
    )
    assert r.status_code == 201, r.text
    temp_pw = r.json()["temporary_password"]

    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"ivan@{tenant}", "password": temp_pw},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    user_tok = r.json()["access_token"]

    r = requests.get(
        f"{api}/api/v1/internal/userinfo",
        headers={**headers_service, "Authorization": f"Bearer {user_tok}"},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    body = r.json()
    assert body["tenant_id"] == tenant
    assert "engineer" in body["roles"]
