"""Сценарий 1: суперадмин создаёт и удаляет предприятие."""
import requests


def test_tenant_lifecycle(api, headers_test):
    r = requests.post(
        f"{api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=30,
    )
    assert r.status_code == 200, r.text
    token = r.json()["access_token"]
    h = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}

    code = "test_lc_acme"
    r = requests.post(f"{api}/api/v1/tenants", headers=h, json={"code": code, "name": "ACME Test"}, timeout=30)
    assert r.status_code == 201, r.text

    r = requests.get(f"{api}/api/v1/tenants", headers=h, timeout=30)
    assert r.status_code == 200
    codes = [t["code"] for t in r.json()]
    assert code in codes

    r = requests.delete(f"{api}/api/v1/tenants/{code}", headers=h, timeout=30)
    assert r.status_code == 204
