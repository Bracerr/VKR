"""E2e tests for RAG fixtures (requires make seed-rag)."""

import json
import os
from pathlib import Path

import pytest
import requests

from conftest import SED_URL, TEST_SECRET

RAG_ENABLED = os.getenv("RAG_FIXTURES_ENABLED", "").lower() == "true"
REPO_ROOT = Path(__file__).resolve().parents[3]
MANIFEST_PATH = REPO_ROOT / "docs" / "rag" / "generated" / "manifest_ids.json"
USERS_PATH = REPO_ROOT / "docs" / "rag" / "generated" / "test_users.json"
AUTH_URL = os.environ.get("AUTH_BASE_URL", "http://localhost:28080")

pytestmark = pytest.mark.skipif(
    not RAG_ENABLED or not MANIFEST_PATH.exists(),
    reason="RAG fixtures not seeded (RAG_FIXTURES_ENABLED=true && make seed-rag)",
)


@pytest.fixture(scope="module")
def manifest():
    return json.loads(MANIFEST_PATH.read_text(encoding="utf-8"))


@pytest.fixture(scope="module")
def rag_users():
    data = json.loads(USERS_PATH.read_text(encoding="utf-8"))
    out = {}
    headers_test = {"X-Test-Secret": TEST_SECRET, "Content-Type": "application/json"}
    for u in data["users"]:
        name = u["username"]
        r = requests.post(
            f"{AUTH_URL}/api/v1/internal/test/login",
            headers=headers_test,
            json={"username": u["login"], "password": u["password"]},
            timeout=60,
        )
        assert r.status_code == 200, f"login {name}: {r.text}"
        out[name] = {
            "headers": {
                "Authorization": f"Bearer {r.json()['access_token']}",
                "Content-Type": "application/json",
            }
        }
    return out


def _ids(h):
    r = requests.get(f"{SED_URL}/api/v1/documents", headers=h, timeout=60)
    assert r.status_code == 200, r.text
    return {d["id"] for d in r.json()}


def test_admin_sees_50_documents(rag_users, manifest):
    ids = _ids(rag_users["rag_admin"]["headers"])
    assert len(ids) == 50
    assert ids == set(manifest["all_document_ids"])


def test_finance_scope(rag_users, manifest):
    ids = _ids(rag_users["rag_finance"]["headers"])
    assert len(ids) == 15
    proc_ids = set()
    for code in ("PURCHASE_REQUEST_APPROVAL", "PURCHASE_ORDER_APPROVAL", "SUPPLIER_CONTRACT_APPROVAL"):
        proc_ids.update(manifest["document_ids_by_type"].get(code, []))
    assert ids == proc_ids


def test_sales_scope(rag_users, manifest):
    ids = _ids(rag_users["rag_sales"]["headers"])
    assert len(ids) == 10


def test_production_scope(rag_users):
    ids = _ids(rag_users["rag_production"]["headers"])
    assert len(ids) == 10


def test_warehouse_scope(rag_users):
    ids = _ids(rag_users["rag_warehouse"]["headers"])
    assert len(ids) == 15


def test_finance_cannot_read_so(rag_users, manifest):
    so_id = manifest["sample_so_doc_id"]
    r = requests.get(f"{SED_URL}/api/v1/documents/{so_id}", headers=rag_users["rag_finance"]["headers"], timeout=60)
    assert r.status_code == 403


def test_finance_can_read_pr(rag_users, manifest):
    pr_id = manifest["sample_pr_doc_id"]
    r = requests.get(f"{SED_URL}/api/v1/documents/{pr_id}", headers=rag_users["rag_finance"]["headers"], timeout=60)
    assert r.status_code == 200
    assert "RAG-PR" in r.json().get("title", "")


def test_sales_cannot_read_pr(rag_users, manifest):
    pr_id = manifest["sample_pr_doc_id"]
    r = requests.get(f"{SED_URL}/api/v1/documents/{pr_id}", headers=rag_users["rag_sales"]["headers"], timeout=60)
    assert r.status_code == 403


def test_no_access_forbidden(rag_users):
    r = requests.get(
        f"{SED_URL}/api/v1/documents",
        headers=rag_users["rag_no_access"]["headers"],
        timeout=60,
    )
    assert r.status_code == 403


def test_author_proc_own_draft(rag_users, manifest):
    draft_id = manifest.get("draft_doc_id")
    if not draft_id:
        pytest.skip("no draft in manifest")
    r = requests.get(
        f"{SED_URL}/api/v1/documents/{draft_id}",
        headers=rag_users["rag_author_proc"]["headers"],
        timeout=60,
    )
    assert r.status_code == 200


def test_author_sales_foreign_pr_forbidden(rag_users, manifest):
    pr_id = manifest["sample_pr_doc_id"]
    r = requests.get(
        f"{SED_URL}/api/v1/documents/{pr_id}",
        headers=rag_users["rag_author_sales"]["headers"],
        timeout=60,
    )
    assert r.status_code == 403


def test_create_document_proc(rag_users, manifest):
    type_id = manifest["type_ids_by_code"]["PURCHASE_REQUEST_APPROVAL"]
    r = requests.post(
        f"{SED_URL}/api/v1/documents",
        headers=rag_users["rag_author_proc"]["headers"],
        json={"type_id": type_id, "title": "RAG e2e create", "payload": {"description": "tmp"}},
        timeout=60,
    )
    assert r.status_code == 201


def test_create_denied_finance(rag_users, manifest):
    type_id = manifest["type_ids_by_code"]["PURCHASE_REQUEST_APPROVAL"]
    r = requests.post(
        f"{SED_URL}/api/v1/documents",
        headers=rag_users["rag_finance"]["headers"],
        json={"type_id": type_id, "title": "denied", "payload": {}},
        timeout=60,
    )
    assert r.status_code == 403


def test_approver_tasks_and_read_in_review(rag_users, manifest):
    in_review = manifest.get("in_review_doc_id")
    if not in_review:
        pytest.skip("no IN_REVIEW doc")
    r = requests.get(f"{SED_URL}/api/v1/tasks", headers=rag_users["rag_approver"]["headers"], timeout=60)
    assert r.status_code == 200
    task_ids = {d["id"] for d in r.json()}
    assert in_review in task_ids
    r2 = requests.get(
        f"{SED_URL}/api/v1/documents/{in_review}",
        headers=rag_users["rag_approver"]["headers"],
        timeout=60,
    )
    assert r2.status_code == 200


def test_history_acl(rag_users, manifest):
    pr_id = manifest["sample_pr_doc_id"]
    so_id = manifest["sample_so_doc_id"]
    h = rag_users["rag_finance"]["headers"]
    assert requests.get(f"{SED_URL}/api/v1/documents/{pr_id}/history", headers=h, timeout=60).status_code == 200
    assert requests.get(f"{SED_URL}/api/v1/documents/{so_id}/history", headers=h, timeout=60).status_code == 403


def test_list_filter_draft_admin(rag_users):
    h = rag_users["rag_admin"]["headers"]
    r = requests.get(f"{SED_URL}/api/v1/documents", headers=h, params={"status": "DRAFT"}, timeout=60)
    assert r.status_code == 200
    assert all(d["status"] == "DRAFT" for d in r.json())
    assert len(r.json()) >= 1
