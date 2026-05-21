#!/usr/bin/env python3
"""Seed RAG test fixtures: 50 SED documents, rag_* users, manifests."""

from __future__ import annotations

import argparse
import json
import os
import sys
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

# scripts/lib
sys.path.insert(0, str(Path(__file__).resolve().parent))
from lib.http_api import ApiError, find_by_code, login_test, pick, req_json  # noqa: E402


def req_json_retry(*args, retries: int = 12, **kwargs):
    for attempt in range(retries):
        try:
            return req_json(*args, **kwargs)
        except ApiError as e:
            if "429" in str(e) and attempt < retries - 1:
                time.sleep(2.0 + attempt * 1.0)
                continue
            raise

TENANT_DEFAULT = "ragcorp"
ENT_ADMIN_USER = "rag_ent_admin"
ENT_ADMIN_PASS_ENV = "RAG_FIXTURES_PASSWORD"

DOC_TYPES: list[tuple[str, str, str, list[str], list[str]]] = [
    # code, name, warehouse_action, reader_roles, writer_roles
    (
        "PURCHASE_REQUEST_APPROVAL",
        "RAG PR Approval",
        "NONE",
        ["doc_read_procurement", "doc_read_finance"],
        ["doc_write_procurement"],
    ),
    (
        "PURCHASE_ORDER_APPROVAL",
        "RAG PO Approval",
        "NONE",
        ["doc_read_procurement", "doc_read_finance"],
        ["doc_write_procurement"],
    ),
    (
        "SUPPLIER_CONTRACT_APPROVAL",
        "RAG Supplier Contract",
        "NONE",
        ["doc_read_procurement", "doc_read_finance"],
        ["doc_write_procurement"],
    ),
    ("SALES_ORDER_APPROVAL", "RAG SO Approval", "NONE", ["doc_read_sales"], ["doc_write_sales"]),
    ("SHIPMENT_APPROVAL", "RAG Shipment", "NONE", ["doc_read_sales"], ["doc_write_sales"]),
    ("BOM_APPROVAL", "RAG BOM", "NONE", ["doc_read_production"], ["doc_write_production"]),
    ("ROUTING_APPROVAL", "RAG Routing", "NONE", ["doc_read_production"], ["doc_write_production"]),
    ("RAG_WH_RESERVE", "RAG Warehouse Reserve", "RESERVE", ["doc_read_warehouse"], ["sed_author"]),
    ("RAG_WH_CONSUME", "RAG Warehouse Consume", "CONSUME", ["doc_read_warehouse"], ["sed_author"]),
    ("RAG_WH_RECEIPT", "RAG Warehouse Receipt", "RECEIPT", ["doc_read_warehouse"], ["sed_author"]),
]

TYPE_SLUG = {
    "PURCHASE_REQUEST_APPROVAL": "PR",
    "PURCHASE_ORDER_APPROVAL": "PO",
    "SUPPLIER_CONTRACT_APPROVAL": "SC",
    "SALES_ORDER_APPROVAL": "SO",
    "SHIPMENT_APPROVAL": "SH",
    "BOM_APPROVAL": "BM",
    "ROUTING_APPROVAL": "RT",
    "RAG_WH_RESERVE": "WR",
    "RAG_WH_CONSUME": "WC",
    "RAG_WH_RECEIPT": "WT",
}

PROCUREMENT_TYPES = {
    "PURCHASE_REQUEST_APPROVAL",
    "PURCHASE_ORDER_APPROVAL",
    "SUPPLIER_CONTRACT_APPROVAL",
}
SALES_TYPES = {"SALES_ORDER_APPROVAL", "SHIPMENT_APPROVAL"}
PRODUCTION_TYPES = {"BOM_APPROVAL", "ROUTING_APPROVAL"}
WAREHOUSE_TYPES = {"RAG_WH_RESERVE", "RAG_WH_CONSUME", "RAG_WH_RECEIPT"}

RAG_USERS: list[tuple[str, list[str]]] = [
    ("rag_admin", ["sed_admin", "sed_author"]),
    ("rag_finance", ["doc_read_finance"]),
    ("rag_procurement", ["doc_read_procurement"]),
    ("rag_sales", ["doc_read_sales"]),
    ("rag_production", ["doc_read_production"]),
    ("rag_warehouse", ["doc_read_warehouse"]),
    ("rag_author_proc", ["sed_author", "doc_write_procurement"]),
    ("rag_author_sales", ["sed_author", "doc_write_sales"]),
    ("rag_approver", ["sed_approver"]),
    ("rag_no_access", ["sed_viewer"]),
]


@dataclass
class Cfg:
    auth_base: str
    sed_base: str
    wh_base: str
    test_secret: str
    wh_secret: str
    tenant: str
    password: str
    output_dir: Path
    superadmin_user: str = "superadmin"
    superadmin_pass: str = "superadmin"
    force: bool = False


@dataclass
class SeedState:
    type_ids: dict[str, str] = field(default_factory=dict)
    type_meta: dict[str, dict[str, Any]] = field(default_factory=dict)
    documents: list[dict[str, Any]] = field(default_factory=list)
    users: dict[str, dict[str, Any]] = field(default_factory=dict)
    wh_ids: dict[str, str] = field(default_factory=dict)
    manifest_ids: dict[str, Any] = field(default_factory=dict)


def load_cfg(args: argparse.Namespace) -> Cfg:
    enabled = os.getenv("RAG_FIXTURES_ENABLED", "true").lower() == "true"
    if not enabled and not args.force:
        print("RAG_FIXTURES_ENABLED is not true; skip seed (use --force to override).")
        sys.exit(0)
    return Cfg(
        auth_base=args.auth_base.rstrip("/"),
        sed_base=args.sed_base.rstrip("/"),
        wh_base=args.wh_base.rstrip("/"),
        test_secret=os.getenv("AUTH_TEST_SECRET", os.getenv("TEST_SECRET", "e2e-test-secret")),
        wh_secret=os.getenv("WAREHOUSE_SERVICE_SECRET", "sed-e2e-wh-secret"),
        tenant=os.getenv("RAG_FIXTURES_TENANT", TENANT_DEFAULT),
        password=os.getenv("RAG_FIXTURES_PASSWORD", "RagTest2026!"),
        output_dir=Path(os.getenv("RAG_FIXTURES_OUTPUT", "docs/rag/generated")),
        force=args.force,
    )


def wh_headers(cfg: Cfg) -> dict[str, str]:
    return {
        "X-Service-Secret": cfg.wh_secret,
        "X-Tenant-Id": cfg.tenant,
        "Content-Type": "application/json",
    }


def ensure_tenant(cfg: Cfg, super_tok: str) -> str:
    tenants = req_json_retry("GET", f"{cfg.auth_base}/api/v1/tenants", token=super_tok, expected=(200,)) or []
    if find_by_code(tenants, cfg.tenant):
        print(f"tenant {cfg.tenant} exists")
    else:
        req_json_retry(
            "POST",
            f"{cfg.auth_base}/api/v1/tenants",
            token=super_tok,
            body={"code": cfg.tenant, "name": f"RAG fixtures {cfg.tenant}"},
            expected=(201,),
        )
        print(f"created tenant {cfg.tenant}")

    try:
        req_json_retry(
            "POST",
            f"{cfg.auth_base}/api/v1/tenants/{cfg.tenant}/ent-admin",
            token=super_tok,
            body={
                "username": ENT_ADMIN_USER,
                "email": f"{ENT_ADMIN_USER}@{cfg.tenant}.test.local",
                "password": cfg.password,
            },
            expected=(201,),
        )
        print(f"created ent-admin {ENT_ADMIN_USER}")
    except ApiError as e:
        if "409" not in str(e) and "exists" not in str(e).lower():
            raise
        print(f"ent-admin {ENT_ADMIN_USER} already exists")

    ent_login = f"{ENT_ADMIN_USER}@{cfg.tenant}"
    try:
        return login_test(cfg.auth_base, cfg.test_secret, ent_login, cfg.password)
    except ApiError:
        # ent-admin мог остаться с временным паролем — пересоздаём через cleanup rag
        req_json_retry(
            "DELETE",
            f"{cfg.auth_base}/api/v1/internal/test/cleanup?prefix=rag",
            headers={"X-Test-Secret": cfg.test_secret},
            expected=(204,),
        )
        req_json_retry(
            "POST",
            f"{cfg.auth_base}/api/v1/tenants/{cfg.tenant}/ent-admin",
            token=super_tok,
            body={
                "username": ENT_ADMIN_USER,
                "email": f"{ENT_ADMIN_USER}@{cfg.tenant}.test.local",
                "password": cfg.password,
            },
            expected=(201,),
        )
        return login_test(cfg.auth_base, cfg.test_secret, ent_login, cfg.password)


def ensure_rag_users(cfg: Cfg, ent_tok: str, st: SeedState) -> None:
    for username, roles in RAG_USERS:
        uname = f"{username}@{cfg.tenant}"
        users = req_json_retry("GET", f"{cfg.auth_base}/api/v1/users", token=ent_tok, expected=(200,)) or []
        existing = next(
            (u for u in users if (u.get("username") or "").startswith(username)),
            None,
        )
        user_pw = cfg.password
        if existing:
            uid = existing.get("keycloak_id") or pick(existing, "id")
        else:
            try:
                r = req_json_retry(
                    "POST",
                    f"{cfg.auth_base}/api/v1/users",
                    token=ent_tok,
                body={
                    "username": username,
                    "email": f"{username}-{cfg.tenant}@test.local",
                    "role": roles[0],
                },
                    expected=(201,),
                )
                uid = r.get("id") or r.get("keycloak_id")
                user_pw = r["temporary_password"]
            except ApiError as e:
                if "409" not in str(e):
                    raise
                users = req_json_retry("GET", f"{cfg.auth_base}/api/v1/users", token=ent_tok, expected=(200,)) or []
                existing = next(
                    (u for u in users if (u.get("username") or "").split("@")[0] == username),
                    None,
                )
                if not existing:
                    raise
                uid = existing.get("keycloak_id") or pick(existing, "id")
        req_json_retry(
            "PUT",
            f"{cfg.auth_base}/api/v1/users/{uid}/roles",
            token=ent_tok,
            body={"roles": roles},
            expected=(204,),
        )
        tok = login_test(cfg.auth_base, cfg.test_secret, uname, user_pw)
        st.users[username] = {
            "username": username,
            "login": uname,
            "id": uid,
            "roles": roles,
            "password": user_pw,
            "token": tok,
        }


def ensure_warehouse(cfg: Cfg, st: SeedState) -> None:
    try:
        _ensure_warehouse_inner(cfg, st)
    except ApiError as e:
        print(f"WARN warehouse seed skipped: {e}")


def _ensure_warehouse_inner(cfg: Cfg, st: SeedState) -> None:
    h = wh_headers(cfg)
    whs = req_json_retry("GET", f"{cfg.wh_base}/api/v1/warehouses", headers=h, expected=(200,)) or []
    wh = find_by_code(whs, "RAG-WH")
    if not wh:
        wh = req_json_retry(
            "POST",
            f"{cfg.wh_base}/api/v1/warehouses",
            headers=h,
            body={"code": "RAG-WH", "name": "RAG Warehouse"},
            expected=(201,),
        )
    wh_id = pick(wh, "id")
    bins = req_json_retry("GET", f"{cfg.wh_base}/api/v1/warehouses/{wh_id}/bins", headers=h, expected=(200,)) or []
    bin_item = find_by_code(bins, "RAG-BIN")
    if not bin_item:
        bin_item = req_json_retry(
            "POST",
            f"{cfg.wh_base}/api/v1/warehouses/{wh_id}/bins",
            headers=h,
            body={"code": "RAG-BIN", "name": "RAG Bin", "bin_type": "STORAGE"},
            expected=(201,),
        )
    bin_id = pick(bin_item, "id")
    prods = req_json_retry("GET", f"{cfg.wh_base}/api/v1/products", headers=h, expected=(200,)) or []
    prod = next((p for p in prods if p.get("sku") == "RAG-SKU-1"), None)
    if not prod:
        try:
            prod = req_json_retry(
                "POST",
                f"{cfg.wh_base}/api/v1/products",
                headers=h,
                body={"sku": "RAG-SKU-1", "name": "RAG Product", "unit": "pcs", "tracking_mode": "NONE"},
                expected=(201,),
            )
        except ApiError as e:
            if "409" not in str(e):
                raise
            prods = req_json_retry("GET", f"{cfg.wh_base}/api/v1/products", headers=h, expected=(200,)) or []
            prod = next((p for p in prods if p.get("sku") == "RAG-SKU-1"), None)
            if not prod:
                raise
    prod_id = pick(prod, "id")
    try:
        req_json_retry(
            "POST",
            f"{cfg.wh_base}/api/v1/operations/receipt",
            headers=h,
            body={
                "warehouse_id": wh_id,
                "bin_id": bin_id,
                "lines": [{"product_id": prod_id, "qty": "500", "unit_cost": "1"}],
            },
            expected=(201,),
        )
    except ApiError:
        pass
    st.wh_ids = {"warehouse_id": wh_id, "bin_id": bin_id, "product_id": prod_id}


def ensure_catalog(cfg: Cfg, admin_tok: str, st: SeedState) -> str:
    wfs = req_json_retry("GET", f"{cfg.sed_base}/api/v1/workflows", token=admin_tok, expected=(200,)) or []
    wf = find_by_code(wfs, "WF-RAG")
    if not wf:
        wf = req_json_retry(
            "POST",
            f"{cfg.sed_base}/api/v1/workflows",
            token=admin_tok,
            body={"code": "WF-RAG", "name": "RAG Approval"},
            expected=(201,),
        )
    wf_id = pick(wf, "id")
    steps = req_json_retry("GET", f"{cfg.sed_base}/api/v1/workflows/{wf_id}/steps", token=admin_tok, expected=(200,)) or []
    if not any(s.get("required_role") == "sed_approver" for s in steps):
        req_json_retry(
            "POST",
            f"{cfg.sed_base}/api/v1/workflows/{wf_id}/steps",
            token=admin_tok,
            body={"order_no": 1, "name": "Approve", "required_role": "sed_approver"},
            expected=(201,),
        )

    dtypes = req_json_retry("GET", f"{cfg.sed_base}/api/v1/document-types", token=admin_tok, expected=(200,)) or []
    for code, name, wh_action, readers, writers in DOC_TYPES:
        dt = find_by_code(dtypes, code)
        if not dt:
            dt = req_json_retry(
                "POST",
                f"{cfg.sed_base}/api/v1/document-types",
                token=admin_tok,
                body={
                    "code": code,
                    "name": name,
                    "warehouse_action": wh_action,
                    "default_workflow_id": wf_id,
                    "reader_roles": readers,
                    "writer_roles": writers,
                },
                expected=(201,),
            )
        tid = pick(dt, "id")
        st.type_ids[code] = tid
        st.type_meta[code] = {
            "id": tid,
            "code": code,
            "warehouse_action": wh_action,
            "reader_roles": readers,
            "writer_roles": writers,
        }
    return wf_id


def author_for_type(code: str) -> str:
    if code in PROCUREMENT_TYPES:
        return "rag_author_proc"
    if code in SALES_TYPES:
        return "rag_author_sales"
    return "rag_admin"


def doc_payload(cfg: Cfg, st: SeedState, code: str, slug: str, idx: int) -> dict[str, Any]:
    rag_id = f"RAG-{slug}-{idx:03d}"
    base = {
        "rag_id": rag_id,
        "description": (
            f"Тестовый документ {rag_id} для индексации RAG. "
            f"Тип {code}. Ключевые слова: закупки продажи склад производство."
        ),
        "keywords": [slug.lower(), "rag", "fixture", f"doc-{idx}"],
    }
    if code in WAREHOUSE_TYPES and st.wh_ids:
        base.update(
            {
                "warehouse_id": st.wh_ids["warehouse_id"],
                "default_bin_id": st.wh_ids["bin_id"],
                "lines": [
                    {
                        "product_id": st.wh_ids["product_id"],
                        "qty": "2",
                        "reason": rag_id,
                        "doc_ref": rag_id,
                    }
                ],
            }
        )
    return base


def advance_document(cfg: Cfg, tok: str, doc_id: str, target: str, approver_tok: str) -> None:
    if target == "DRAFT":
        return
    time.sleep(0.6)
    req_json_retry("POST", f"{cfg.sed_base}/api/v1/documents/{doc_id}/submit", token=tok, expected=(204,))
    if target == "IN_REVIEW":
        return
    time.sleep(0.6)
    req_json_retry(
        "POST",
        f"{cfg.sed_base}/api/v1/documents/{doc_id}/approve",
        token=approver_tok,
        body={"comment": "rag seed approve"},
        expected=(204,),
    )
    if target == "SIGNED":
        time.sleep(0.6)
        req_json_retry("POST", f"{cfg.sed_base}/api/v1/documents/{doc_id}/sign", token=tok, expected=(204,))


def create_documents(cfg: Cfg, st: SeedState) -> None:
    admin_tok = st.users["rag_admin"]["token"]
    approver_tok = st.users["rag_approver"]["token"]
    existing = req_json_retry("GET", f"{cfg.sed_base}/api/v1/documents", token=admin_tok, expected=(200,)) or []
    rag_existing = [d for d in existing if "RAG-" in (d.get("title") or "")]
    if len(rag_existing) >= 50:
        print(f"found {len(rag_existing)} RAG documents, skip creation")
        for d in rag_existing:
            st.documents.append(normalize_doc(d, st))
        build_manifest_ids(st)
        return

    targets = ["APPROVED", "APPROVED", "APPROVED", "DRAFT", "IN_REVIEW"]
    for code, _name, wh_action, readers, writers in DOC_TYPES:
        slug = TYPE_SLUG[code]
        author_key = author_for_type(code)
        author_tok = st.users[author_key]["token"]
        for i in range(1, 6):
            rag_id = f"RAG-{slug}-{i:03d}"
            target = targets[i - 1]
            if code == "RAG_WH_RESERVE" and i == 1 and st.wh_ids:
                target = "SIGNED"
            title = f"{rag_id}: {code}"
            body = {
                "type_id": st.type_ids[code],
                "title": title,
                "payload": doc_payload(cfg, st, code, slug, i),
            }
            time.sleep(0.5)
            doc = req_json_retry(
                "POST",
                f"{cfg.sed_base}/api/v1/documents",
                token=author_tok,
                body=body,
                expected=(201,),
            )
            doc_id = pick(doc, "id")
            advance_document(cfg, author_tok, doc_id, target, approver_tok)
            full = req_json_retry(
                "GET",
                f"{cfg.sed_base}/api/v1/documents/{doc_id}",
                token=admin_tok,
                expected=(200,),
            )
            entry = normalize_doc(full, st)
            entry["rag_id"] = rag_id
            entry["target_status"] = target
            st.documents.append(entry)
    build_manifest_ids(st)


def normalize_doc(d: dict[str, Any], st: SeedState) -> dict[str, Any]:
    type_id = pick(d, "type_id")
    type_code = next((c for c, tid in st.type_ids.items() if tid == type_id), "")
    meta = st.type_meta.get(type_code, {})
    payload = d.get("payload") or {}
    if isinstance(payload, str):
        try:
            payload = json.loads(payload)
        except json.JSONDecodeError:
            payload = {"raw": payload}
    desc = payload.get("description", "") if isinstance(payload, dict) else ""
    title = d.get("title") or ""
    return {
        "document_id": pick(d, "id"),
        "type_id": type_id,
        "type_code": type_code,
        "title": title,
        "status": d.get("status"),
        "author_sub": d.get("author_sub"),
        "payload": payload,
        "reader_roles": meta.get("reader_roles", []),
        "writer_roles": meta.get("writer_roles", []),
        "search_text": f"{title}\n{desc}",
    }


def build_manifest_ids(st: SeedState) -> None:
    by_type: dict[str, list[str]] = {}
    for doc in st.documents:
        by_type.setdefault(doc["type_code"], []).append(doc["document_id"])
    draft_id = next((d["document_id"] for d in st.documents if d.get("status") == "DRAFT"), None)
    in_review_id = next((d["document_id"] for d in st.documents if d.get("status") == "IN_REVIEW"), None)
    pr_id = None
    so_id = None
    for d in st.documents:
        if d["type_code"] == "PURCHASE_REQUEST_APPROVAL" and pr_id is None:
            pr_id = d["document_id"]
        if d["type_code"] == "SALES_ORDER_APPROVAL" and so_id is None:
            so_id = d["document_id"]
    st.manifest_ids = {
        "tenant": st.users.get("_tenant", ""),
        "type_ids_by_code": st.type_ids,
        "document_ids_by_type": by_type,
        "all_document_ids": [d["document_id"] for d in st.documents],
        "draft_doc_id": draft_id,
        "in_review_doc_id": in_review_id,
        "sample_pr_doc_id": pr_id,
        "sample_so_doc_id": so_id,
        "users": {
            k: {"login": v["login"], "roles": v["roles"], "sub_hint": v.get("login")}
            for k, v in st.users.items()
            if isinstance(v, dict) and "login" in v
        },
    }


def roles_overlap(user_roles: list[str], allowed: list[str]) -> bool:
    return bool(set(user_roles) & set(allowed))


def expected_visible(doc: dict[str, Any], username: str, roles: list[str], sub: str) -> bool:
    if "sed_admin" in roles:
        return True
    if doc.get("author_sub") == sub:
        return True
    return roles_overlap(roles, doc.get("reader_roles") or [])


def compute_access_matrix(st: SeedState) -> list[dict[str, Any]]:
    matrix = []
    for username, u in st.users.items():
        roles = u["roles"]
        login = u["login"]
        visible = []
        reasons: dict[str, str] = {}
        for doc in st.documents:
            sub = login  # test login username@tenant maps to sub in KC
            if "sed_admin" in roles:
                visible.append(doc["document_id"])
                reasons[doc["document_id"]] = "admin"
            elif doc.get("author_sub") and username in ("rag_author_proc", "rag_author_sales", "rag_admin"):
                if doc["document_id"] not in visible and doc.get("status") == "DRAFT":
                    if (username == "rag_author_proc" and doc["type_code"] in PROCUREMENT_TYPES) or (
                        username == "rag_author_sales" and doc["type_code"] in SALES_TYPES
                    ):
                        visible.append(doc["document_id"])
                        reasons[doc["document_id"]] = "author"
            if roles_overlap(roles, doc.get("reader_roles") or []):
                if doc["document_id"] not in visible:
                    visible.append(doc["document_id"])
                    reasons[doc["document_id"]] = "reader_role"
        matrix.append(
            {
                "username": username,
                "login": login,
                "roles": roles,
                "visible_document_ids": visible,
                "visibility_reasons": reasons,
                "expected_count": len(visible),
            }
        )
    return matrix


EXPECTED_COUNTS = {
    "rag_admin": 50,
    "rag_finance": 15,
    "rag_procurement": 15,
    "rag_sales": 10,
    "rag_production": 10,
    "rag_warehouse": 15,
    "rag_no_access": 0,
}


def verify_access_api(cfg: Cfg, st: SeedState) -> dict[str, Any]:
    mismatches: list[str] = []
    matrix = compute_access_matrix(st)
    for entry in matrix:
        username = entry["username"]
        if username not in EXPECTED_COUNTS or username == "rag_no_access":
            continue
        tok = st.users[username]["token"]
        if username == "rag_no_access":
            try:
                req_json_retry("GET", f"{cfg.sed_base}/api/v1/documents", token=tok, expected=(200,))
                docs = []
            except ApiError as e:
                if "403" not in str(e):
                    raise
                docs = []
        else:
            docs = req_json_retry("GET", f"{cfg.sed_base}/api/v1/documents", token=tok, expected=(200,)) or []
        api_ids = {pick(d, "id") for d in docs}
        exp = EXPECTED_COUNTS[username]
        if len(api_ids) != exp:
            mismatches.append(f"{username}: api={len(api_ids)} expected={exp}")
    return {"mismatches": mismatches, "matrix": matrix, "api_counts": {e["username"]: len(e.get("visible_document_ids", [])) for e in matrix}}


def write_manifests(cfg: Cfg, st: SeedState, report: dict[str, Any]) -> None:
    out = cfg.output_dir
    out.mkdir(parents=True, exist_ok=True)
    st.users["_tenant"] = cfg.tenant
    build_manifest_ids(st)
    (out / "corpus_full.json").write_text(
        json.dumps({"tenant": cfg.tenant, "documents": st.documents}, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    (out / "document_types.json").write_text(
        json.dumps(list(st.type_meta.values()), ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    user_list = [
        {k: v for k, v in u.items() if k != "token"}
        for uname, u in st.users.items()
        if not uname.startswith("_")
    ]
    (out / "test_users.json").write_text(
        json.dumps(
            {"tenant": cfg.tenant, "password": cfg.password, "users": user_list},
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    (out / "access_matrix.json").write_text(
        json.dumps(report.get("matrix", []), ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    (out / "verification_report.json").write_text(
        json.dumps(report, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    (out / "manifest_ids.json").write_text(
        json.dumps(st.manifest_ids, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    print(f"wrote manifests to {out}")


def _purge_sed_tenant_documents(tenant: str) -> None:
    """Удаляет документы тенанта в e2e Postgres (нет HTTP API delete)."""
    import subprocess

    sql = f"DELETE FROM documents WHERE tenant_code = '{tenant}';"
    for container in ("vkr-test-pg-sed-1", "pg-sed-1"):
        try:
            subprocess.run(
                ["docker", "exec", container, "psql", "-U", "sed", "-d", "sed", "-c", sql],
                check=True,
                capture_output=True,
                timeout=30,
            )
            print(f"purged SED documents for tenant {tenant} via {container}")
            return
        except (subprocess.CalledProcessError, FileNotFoundError, subprocess.TimeoutExpired):
            continue
    print(f"WARN: could not purge SED documents for {tenant} (docker psql)")


def teardown(cfg: Cfg) -> None:
    _purge_sed_tenant_documents(cfg.tenant)
    h = {"X-Test-Secret": cfg.test_secret}
    try:
        req_json_retry(
            "DELETE",
            f"{cfg.auth_base}/api/v1/internal/test/cleanup?prefix=rag",
            headers=h,
            expected=(204,),
        )
        print("cleaned Keycloak users with prefix rag")
    except ApiError as e:
        print(f"cleanup rag users: {e}")
    super_tok = login_test(cfg.auth_base, cfg.test_secret, cfg.superadmin_user, cfg.superadmin_pass)
    try:
        req_json_retry(
            "DELETE",
            f"{cfg.auth_base}/api/v1/tenants/{cfg.tenant}",
            token=super_tok,
            expected=(204, 404),
        )
        print(f"deleted tenant {cfg.tenant}")
    except ApiError as e:
        print(f"teardown: {e}")
    out = cfg.output_dir
    for name in (
        "corpus_full.json",
        "access_matrix.json",
        "test_users.json",
        "document_types.json",
        "verification_report.json",
        "manifest_ids.json",
    ):
        p = out / name
        if p.exists():
            p.unlink()
    print("removed generated manifests")


def run_seed(cfg: Cfg) -> int:
    time.sleep(5)  # cooldown after prior e2e / partial seeds
    st = SeedState()
    super_tok = login_test(cfg.auth_base, cfg.test_secret, cfg.superadmin_user, cfg.superadmin_pass)
    ent_tok = ensure_tenant(cfg, super_tok)
    ensure_rag_users(cfg, ent_tok, st)
    for u in st.users.values():
        u["token"] = login_test(cfg.auth_base, cfg.test_secret, u["login"], u["password"])
    ensure_warehouse(cfg, st)
    admin_tok = st.users["rag_admin"]["token"]
    ensure_catalog(cfg, admin_tok, st)
    create_documents(cfg, st)
    report = verify_access_api(cfg, st)
    if report["mismatches"]:
        print("WARN access mismatches:", report["mismatches"])
    write_manifests(cfg, st, report)
    print(f"RAG seed done: {len(st.documents)} documents, tenant={cfg.tenant}")
    return 0


def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Seed RAG test fixtures")
    p.add_argument("--auth-base", default=os.getenv("AUTH_BASE_URL", "http://localhost:28080"))
    p.add_argument("--sed-base", default=os.getenv("SED_BASE_URL", "http://localhost:28091"))
    p.add_argument("--wh-base", default=os.getenv("WAREHOUSE_BASE_URL", "http://localhost:28090"))
    p.add_argument("--teardown", action="store_true")
    p.add_argument("--force", action="store_true")
    return p.parse_args()


def main() -> int:
    args = parse_args()
    cfg = load_cfg(args)
    try:
        if args.teardown:
            teardown(cfg)
            return 0
        return run_seed(cfg)
    except ApiError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
