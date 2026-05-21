"""Warehouse sign smoke for RAG WH_RESERVE fixture."""

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
    reason="RAG fixtures not seeded",
)


@pytest.fixture(scope="module")
def admin_headers():
    data = json.loads(USERS_PATH.read_text(encoding="utf-8"))
    admin = next(u for u in data["users"] if u["username"] == "rag_admin")
    r = requests.post(
        f"{AUTH_URL}/api/v1/internal/test/login",
        headers={"X-Test-Secret": TEST_SECRET, "Content-Type": "application/json"},
        json={"username": admin["login"], "password": admin["password"]},
        timeout=60,
    )
    assert r.status_code == 200
    return {"Authorization": f"Bearer {r.json()['access_token']}", "Content-Type": "application/json"}


def test_wh_reserve_signed_exists(admin_headers):
    r = requests.get(f"{SED_URL}/api/v1/documents", headers=admin_headers, timeout=60)
    assert r.status_code == 200
    signed_wh = [
        d
        for d in r.json()
        if "RAG-WR-001" in (d.get("title") or "") and d.get("status") == "SIGNED"
    ]
    if not signed_wh:
        pytest.skip("RAG-WR-001 not signed (warehouse seed optional)")
    doc = signed_wh[0]
    assert doc.get("warehouse_ref")
