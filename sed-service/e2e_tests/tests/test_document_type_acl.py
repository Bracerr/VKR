"""ACL по типам документов: doc_read_finance vs doc_read_sales."""
import uuid

import requests


def test_document_type_acl(auth_api, sed_api, headers_test):
    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": "superadmin", "password": "superadmin"},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    h_super = {"Authorization": f"Bearer {r.json()['access_token']}", "Content-Type": "application/json"}

    tenant = "test_acl_" + uuid.uuid4().hex[:10]
    r = requests.post(
        f"{auth_api}/api/v1/tenants",
        headers=h_super,
        json={"code": tenant, "name": "ACL E2E"},
        timeout=60,
    )
    assert r.status_code == 201, r.text

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
    h_ent = {"Authorization": f"Bearer {r.json()['access_token']}", "Content-Type": "application/json"}

    r = requests.post(
        f"{auth_api}/api/v1/users",
        headers=h_ent,
        json={"username": "sedadmin", "email": f"sedadmin-{tenant}@test.local", "role": "sed_admin"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    admin_uid = r.json()["id"]
    admin_pw = r.json()["temporary_password"]

    admin_roles = [
        "sed_admin",
        "sed_author",
        "doc_read_procurement",
        "doc_read_sales",
        "doc_write_procurement",
        "doc_write_sales",
    ]
    r = requests.put(
        f"{auth_api}/api/v1/users/{admin_uid}/roles",
        headers=h_ent,
        json={"roles": admin_roles},
        timeout=60,
    )
    assert r.status_code == 204, r.text

    r = requests.post(
        f"{auth_api}/api/v1/internal/test/login",
        headers=headers_test,
        json={"username": f"sedadmin@{tenant}", "password": admin_pw},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    h_admin = {"Authorization": f"Bearer {r.json()['access_token']}", "Content-Type": "application/json"}

    r = requests.post(
        f"{sed_api}/api/v1/workflows",
        headers=h_admin,
        json={"code": "WF-ACL", "name": "ACL"},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    wf_id = r.json()["id"]

    pr_body = {
        "code": "PURCHASE_REQUEST_APPROVAL",
        "name": "PR ACL",
        "warehouse_action": "NONE",
        "default_workflow_id": wf_id,
        "reader_roles": ["doc_read_finance", "doc_read_procurement"],
        "writer_roles": ["doc_write_procurement"],
    }
    r = requests.post(f"{sed_api}/api/v1/document-types", headers=h_admin, json=pr_body, timeout=60)
    assert r.status_code == 201, r.text
    pr_type_id = r.json()["id"]

    so_body = {
        "code": "SALES_ORDER_APPROVAL",
        "name": "SO ACL",
        "warehouse_action": "NONE",
        "default_workflow_id": wf_id,
        "reader_roles": ["doc_read_sales"],
        "writer_roles": ["doc_write_sales"],
    }
    r = requests.post(f"{sed_api}/api/v1/document-types", headers=h_admin, json=so_body, timeout=60)
    assert r.status_code == 201, r.text
    so_type_id = r.json()["id"]

    r = requests.post(
        f"{sed_api}/api/v1/documents",
        headers=h_admin,
        json={"type_id": pr_type_id, "title": "PR doc", "payload": {}},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    pr_doc_id = r.json()["id"]

    r = requests.post(
        f"{sed_api}/api/v1/documents",
        headers=h_admin,
        json={"type_id": so_type_id, "title": "SO doc", "payload": {}},
        timeout=60,
    )
    assert r.status_code == 201, r.text
    so_doc_id = r.json()["id"]

    def _mk_reader(username: str, roles: list[str]):
        r = requests.post(
            f"{auth_api}/api/v1/users",
            headers=h_ent,
            json={"username": username, "email": f"{username}-{tenant}@test.local", "role": "viewer"},
            timeout=60,
        )
        assert r.status_code == 201, r.text
        uid = r.json()["id"]
        pw = r.json()["temporary_password"]
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
            json={"username": f"{username}@{tenant}", "password": pw},
            timeout=60,
        )
        assert r.status_code == 200, r.text
        return {"Authorization": f"Bearer {r.json()['access_token']}", "Content-Type": "application/json"}

    h_fin = _mk_reader("fin", ["doc_read_finance"])
    h_sales = _mk_reader("salesro", ["doc_read_sales"])

    r = requests.get(f"{sed_api}/api/v1/documents/{pr_doc_id}", headers=h_fin, timeout=60)
    assert r.status_code == 200, r.text

    r = requests.get(f"{sed_api}/api/v1/documents/{so_doc_id}", headers=h_fin, timeout=60)
    assert r.status_code == 403, r.text

    r = requests.get(f"{sed_api}/api/v1/documents/{so_doc_id}", headers=h_sales, timeout=60)
    assert r.status_code == 200, r.text

    r = requests.get(f"{sed_api}/api/v1/documents/{pr_doc_id}", headers=h_sales, timeout=60)
    assert r.status_code == 403, r.text

    r = requests.get(f"{sed_api}/api/v1/documents", headers=h_fin, timeout=60)
    assert r.status_code == 200, r.text
    ids = {d["id"] for d in r.json()}
    assert pr_doc_id in ids
    assert so_doc_id not in ids
