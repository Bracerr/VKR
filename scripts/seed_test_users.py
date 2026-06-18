#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import os
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from lib.http_api import ApiError, login_test, pick, req_json  # noqa: E402

POWER_USER_ROLES: list[str] = [
    "ent_admin",
    "approver",
    "engineer",
    "viewer",
    "warehouse_admin",
    "storekeeper",
    "warehouse_viewer",
    "sed_admin",
    "sed_author",
    "sed_approver",
    "sed_viewer",
    "doc_read_procurement",
    "doc_read_sales",
    "doc_read_production",
    "doc_read_warehouse",
    "doc_read_finance",
    "doc_write_procurement",
    "doc_write_sales",
    "doc_write_production",
    "prod_admin",
    "prod_technologist",
    "prod_planner",
    "prod_master",
    "prod_worker",
    "prod_qc",
    "prod_viewer",
    "proc_admin",
    "proc_buyer",
    "proc_approver",
    "proc_viewer",
    "sales_admin",
    "sales_manager",
    "sales_approver",
    "sales_viewer",
]

TEST_USERS: list[tuple[str, list[str], str]] = [
    ("test1", POWER_USER_ROLES, "Полный доступ ко всем модулям тенанта (как vkr_demo)"),
    (
        "test2",
        ["storekeeper", "warehouse_viewer", "doc_read_warehouse"],
        "Склад: операции и чтение остатков, без продаж/закупок",
    ),
    (
        "test3",
        ["sales_manager", "doc_read_sales", "doc_write_sales", "sed_author", "sed_viewer"],
        "Продажи: заказы клиентов и документы продаж, без склада и закупок",
    ),
    (
        "test4",
        ["proc_buyer", "doc_read_procurement", "doc_write_procurement", "sed_approver"],
        "Закупки: PR/PO и согласование, без продаж и склада",
    ),
    (
        "test5",
        ["sed_viewer", "warehouse_viewer", "proc_viewer", "sales_viewer", "prod_viewer"],
        "Только чтение во всех модулях",
    ),
]

ENT_ADMIN_USER = "rag_ent_admin"


def req_retry(*args, retries: int = 8, **kwargs):
    for attempt in range(retries):
        try:
            return req_json(*args, **kwargs)
        except ApiError as e:
            msg = str(e)
            if attempt < retries - 1 and any(x in msg for x in ("429", "500", "502", "503")):
                time.sleep(1.5 + attempt)
                continue
            raise


def ent_admin_token(auth_base: str, test_secret: str, tenant: str, ent_password: str, rag_secret: str) -> str:
    login = f"{ENT_ADMIN_USER}@{tenant}"
    try:
        return login_test(auth_base, test_secret, login, ent_password)
    except ApiError:
        req_retry(
            "POST",
            f"{auth_base}/api/v1/internal/tenants/{tenant}/repair-passwords",
            headers={"X-Service-Secret": rag_secret},
            body={"password": ent_password},
            expected=(200,),
        )
        return login_test(auth_base, test_secret, login, ent_password)


def upsert_user(
    auth_base: str,
    test_secret: str,
    ent_tok: str,
    tenant: str,
    username: str,
    roles: list[str],
    password: str,
) -> dict:
    uname = f"{username}@{tenant}"
    users = req_retry("GET", f"{auth_base}/api/v1/users", token=ent_tok, expected=(200,)) or []
    existing = next((u for u in users if (u.get("username") or "").split("@")[0] == username), None)
    if existing:
        uid = existing.get("keycloak_id") or pick(existing, "id")
    else:
        try:
            r = req_retry(
                "POST",
                f"{auth_base}/api/v1/users",
                token=ent_tok,
                body={
                    "username": username,
                    "email": f"{username}-{tenant}@test.local",
                    "role": roles[0],
                    "password": password,
                },
                expected=(201,),
            )
            uid = r.get("id") or r.get("keycloak_id")
        except ApiError as e:
            if "409" not in str(e):
                raise
            users = req_retry("GET", f"{auth_base}/api/v1/users", token=ent_tok, expected=(200,)) or []
            existing = next((u for u in users if (u.get("username") or "").split("@")[0] == username), None)
            if not existing:
                raise
            uid = existing.get("keycloak_id") or pick(existing, "id")

    req_retry(
        "PUT",
        f"{auth_base}/api/v1/users/{uid}/password",
        token=ent_tok,
        body={"password": password},
        expected=(204,),
    )
    req_retry(
        "PUT",
        f"{auth_base}/api/v1/users/{uid}/roles",
        token=ent_tok,
        body={"roles": roles},
        expected=(204,),
    )
    login_test(auth_base, test_secret, uname, password)
    return {"username": username, "login": uname, "roles": roles, "password": password}


def main() -> None:
    p = argparse.ArgumentParser(description="Seed test users test1..test5 in tenant ragcorp")
    p.add_argument("--auth-base", default=os.getenv("AUTH_BASE_URL", "http://localhost:28080"))
    p.add_argument("--tenant", default=os.getenv("TEST_USERS_TENANT", os.getenv("RAG_FIXTURES_TENANT", "ragcorp")))
    p.add_argument(
        "--password",
        default=os.getenv("TEST_USERS_PASSWORD", "test1234"),
        help="Пароль test-пользователей (мин. 8 символов в Keycloak)",
    )
    p.add_argument(
        "--ent-password",
        default=os.getenv("RAG_FIXTURES_PASSWORD", "RagTest2026!"),
        help="Пароль ent_admin для создания пользователей",
    )
    p.add_argument("--output", default=os.getenv("TEST_USERS_OUTPUT", "docs/test_users.json"))
    args = p.parse_args()

    if len(args.password) < 8:
        print("ERROR: пароль test-пользователей должен быть не короче 8 символов (ограничение Keycloak).", file=sys.stderr)
        sys.exit(1)

    auth_base = args.auth_base.rstrip("/")
    test_secret = os.getenv("AUTH_TEST_SECRET", os.getenv("TEST_SECRET", "e2e-test-secret"))
    rag_secret = os.getenv("RAG_CORPUS_SECRET", os.getenv("AUTH_SERVICE_SECRET", "e2e-service-secret"))

    ent_tok = ent_admin_token(auth_base, test_secret, args.tenant, args.ent_password, rag_secret)

    manifest: dict = {
        "tenant": args.tenant,
        "password": args.password,
        "users": [],
    }
    for username, roles, desc in TEST_USERS:
        info = upsert_user(auth_base, test_secret, ent_tok, args.tenant, username, roles, args.password)
        info["description"] = desc
        manifest["users"].append(info)
        print(f"OK {info['login']} — {desc} ({len(roles)} ролей)")

    out = Path(args.output)
    out.parent.mkdir(parents=True, exist_ok=True)
    out.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"\nМанифест: {out}")
    print(f"Пароль всех test*: {args.password}")


if __name__ == "__main__":
    main()
