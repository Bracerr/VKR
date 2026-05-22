"""Internal RAG corpus: text + access + attachment URLs only."""

import os

import pytest
import requests

SED_URL = os.getenv("SED_URL", "http://localhost:5656").rstrip("/")
RAG_TENANT = os.getenv("RAG_FIXTURES_TENANT", "ragcorp")
RAG_SECRET = os.getenv("RAG_CORPUS_SECRET", os.getenv("AUTH_SERVICE_SECRET", "e2e-service-secret"))
TEST_SECRET = os.getenv("AUTH_TEST_SECRET", os.getenv("TEST_SECRET", "e2e-test-secret"))
RAG_PASSWORD = os.getenv("RAG_FIXTURES_PASSWORD", "RagTest2026!")


def _login(username: str) -> str:
    r = requests.post(
        f"{SED_URL}/api/v1/internal/test/login",
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


def test_public_documents_have_no_rag_index_fields():
    tok = _login("rag_admin")
    r = requests.get(
        f"{SED_URL}/api/v1/documents",
        headers={"Authorization": f"Bearer {tok}"},
        timeout=60,
    )
    assert r.status_code == 200
    for d in r.json():
        if "RAG-" not in (d.get("title") or ""):
            continue
        payload = d.get("payload") or {}
        assert "search_text" not in payload
        assert "rag_id" not in payload


def test_rag_corpus_minimal_shape(rag_corpus):
    assert "documents" in rag_corpus
    assert "users" not in rag_corpus
    assert "tenant" not in rag_corpus
    docs = rag_corpus["documents"]
    assert len(docs) >= 50
    sample = docs[0]
    assert set(sample.keys()) == {"document_id", "text", "access", "attachments"}
    assert "Подшипник" in sample["text"] or "закупк" in sample["text"].lower() or len(sample["text"]) > 20
    access = sample["access"]
    assert "read_roles" in access
    assert "write_roles" in access
    assert "approve_roles" in access
    assert "admin_roles" in access
    assert isinstance(sample["attachments"], list)


def test_rag_corpus_rejects_wrong_secret():
    r = requests.get(
        f"{SED_URL}/api/v1/internal/rag/corpus",
        headers={"X-Service-Secret": "wrong", "X-Tenant-Id": RAG_TENANT},
        timeout=30,
    )
    assert r.status_code == 401
