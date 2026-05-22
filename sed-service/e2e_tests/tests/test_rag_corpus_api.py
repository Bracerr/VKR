"""Internal RAG corpus API и отсутствие RAG-полей в публичном GET /documents."""

import os

import pytest
import requests

SED_URL = os.getenv("SED_URL", "http://localhost:5656").rstrip("/")
AUTH_URL = os.getenv("AUTH_URL", "http://localhost:5656").rstrip("/")
RAG_TENANT = os.getenv("RAG_FIXTURES_TENANT", "ragcorp")
RAG_SECRET = os.getenv("RAG_CORPUS_SECRET", os.getenv("AUTH_SERVICE_SECRET", "e2e-service-secret"))
TEST_SECRET = os.getenv("AUTH_TEST_SECRET", os.getenv("TEST_SECRET", "e2e-test-secret"))
RAG_PASSWORD = os.getenv("RAG_FIXTURES_PASSWORD", "RagTest2026!")


def _login(username: str) -> str:
    r = requests.post(
        f"{AUTH_URL}/api/v1/internal/test/login",
        headers={"X-Test-Secret": TEST_SECRET, "Content-Type": "application/json"},
        json={"username": f"{username}@{RAG_TENANT}", "password": RAG_PASSWORD},
        timeout=60,
    )
    assert r.status_code == 200, r.text
    return r.json()["access_token"]


@pytest.fixture(scope="module")
def rag_corpus():
    r = requests.get(
        f"{SED_URL}/api/v1/internal/rag/corpus",
        headers={"X-Service-Secret": RAG_SECRET, "X-Tenant-Id": RAG_TENANT},
        timeout=120,
    )
    assert r.status_code == 200, r.text
    return r.json()


def test_public_documents_have_no_rag_content():
    tok = _login("rag_admin")
    r = requests.get(
        f"{SED_URL}/api/v1/documents",
        headers={"Authorization": f"Bearer {tok}"},
        timeout=60,
    )
    assert r.status_code == 200
    docs = r.json()
    assert len(docs) >= 50
    for d in docs:
        if "RAG-" not in (d.get("title") or ""):
            continue
        payload = d.get("payload") or {}
        assert "rag_id" not in payload
        assert "search_text" not in payload
        assert "keywords" not in payload or "fixture" not in str(payload.get("keywords", ""))


def test_rag_corpus_returns_documents_and_access(rag_corpus):
    assert rag_corpus.get("tenant") == RAG_TENANT
    docs = rag_corpus.get("documents") or []
    users = rag_corpus.get("users") or []
    assert len(docs) >= 50
    assert len(users) >= 5
    sample = next(d for d in docs if d.get("search_text"))
    assert sample.get("content")
    assert sample.get("reader_roles")
    finance = next((u for u in users if "rag_finance" in (u.get("login") or "")), None)
    assert finance is not None
    assert finance.get("expected_count") == 15


def test_rag_corpus_rejects_wrong_secret():
    r = requests.get(
        f"{SED_URL}/api/v1/internal/rag/corpus",
        headers={"X-Service-Secret": "wrong", "X-Tenant-Id": RAG_TENANT},
        timeout=30,
    )
    assert r.status_code == 401
