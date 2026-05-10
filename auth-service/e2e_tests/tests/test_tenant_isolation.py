"""Сценарий 3: ent_admin тенанта A не может удалить пользователя тенанта B."""
import requests


def _super_token(api, headers_test):
    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    return r.json()["access_token"]


def test_cross_tenant_delete_forbidden(api, headers_test):
    super_tok = _super_token(api, headers_test)
    h_super = {"Authorization": f"Bearer {super_tok}", "Content-Type": "application/json"}

    ta, tb = "test_iso_a", "test_iso_b"
    for code, name in [(ta, "ISO A"), (tb, "ISO B")]:
        requests.delete(f"{api}/api/v1/tenants/{code}", headers=h_super, timeout=30)
        r = requests.post(f"{api}/api/v1/tenants", headers=h_super, json={"code": code, "name": name}, timeout=30)
        assert r.status_code == 201, r.text
        r = requests.post(
            f"{api}/api/v1/tenants/{code}/ent-admin",
            headers=h_super,
            json={"username": "adm", "email": f"adm@{code}.local", "password": "AdminPass123!"},
            timeout=30,
        )
        assert r.status_code == 201, r.text

    # ent_admin A
    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"adm@{ta}", "password": "AdminPass123!"},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    tok_a = r.json()["access_token"]
    h_a = {"Authorization": f"Bearer {tok_a}", "Content-Type": "application/json"}

    # пользователь в B
    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"adm@{tb}", "password": "AdminPass123!"},
        timeout=30,
    )
    tok_b = r.json()["access_token"]
    h_b = {"Authorization": f"Bearer {tok_b}", "Content-Type": "application/json"}
    r = requests.post(
        f"{api}/api/v1/users",
        headers=h_b,
        json={"username": "victim", "email": "v@b.local", "role": "viewer"},
        timeout=30,
    )
    assert r.status_code == 201, r.text
    victim_id = r.json()["id"]

    r = requests.delete(f"{api}/api/v1/users/{victim_id}", headers=h_a, timeout=30)
    assert r.status_code in (403, 404), r.text
