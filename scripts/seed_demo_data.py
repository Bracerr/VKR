#!/usr/bin/env python3
"""Seed demo data across VKR services for dev environment."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import sys
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Any


def now_suffix() -> str:
    return dt.datetime.utcnow().strftime("%Y%m%d%H%M%S")


class ApiError(Exception):
    pass


@dataclass
class Cfg:
    gateway_base: str
    production_direct_base: str
    procurement_direct_base: str
    test_secret: str
    callback_secret: str
    superadmin_username: str
    superadmin_password: str


def req_json(
    method: str,
    url: str,
    *,
    token: str | None = None,
    body: dict[str, Any] | None = None,
    headers: dict[str, str] | None = None,
    expected: tuple[int, ...] = (200,),
) -> Any:
    raw = None
    req_headers = {"Accept": "application/json"}
    if headers:
        req_headers.update(headers)
    if token:
        req_headers["Authorization"] = f"Bearer {token}"
    if body is not None:
        raw = json.dumps(body).encode("utf-8")
        req_headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=raw, method=method, headers=req_headers)
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            status = resp.getcode()
            payload = resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        text = e.read().decode("utf-8", errors="replace")
        raise ApiError(f"{method} {url} -> {e.code}: {text}") from e
    except urllib.error.URLError as e:
        raise ApiError(f"{method} {url} -> network error: {e}") from e

    if status not in expected:
        raise ApiError(f"{method} {url} -> {status}: {payload}")
    if not payload:
        return None
    try:
        return json.loads(payload)
    except json.JSONDecodeError:
        return payload


def login_superadmin(cfg: Cfg) -> str:
    body = {
        "username": cfg.superadmin_username,
        "password": cfg.superadmin_password,
    }
    data = req_json(
        "POST",
        f"{cfg.gateway_base}/api/v1/internal/test/login",
        body=body,
        headers={"X-Test-Secret": cfg.test_secret},
        expected=(200,),
    )
    return data["access_token"]


def find_by_code(items: list[dict[str, Any]], code: str) -> dict[str, Any] | None:
    for item in items:
        if pick(item, "code") == code:
            return item
    return None


def pick(item: dict[str, Any], *keys: str) -> Any:
    def norm(s: str) -> str:
        return s.lower().replace("_", "").replace("-", "")

    lowered = {norm(k): v for k, v in item.items()}
    for k in keys:
        nk = norm(k)
        if nk in lowered:
            return lowered[nk]
    raise KeyError(keys[0] if keys else "key")


def ensure_warehouse_basics(base: str, token: str) -> dict[str, str]:
    whs = req_json("GET", f"{base}/api/v1/warehouses", token=token, expected=(200,)) or []
    wh = find_by_code(whs, "DEMO-WH")
    if not wh:
        wh = req_json(
            "POST",
            f"{base}/api/v1/warehouses",
            token=token,
            body={"code": "DEMO-WH", "name": "Demo Warehouse"},
            expected=(201,),
        )
    wh_id = pick(wh, "id")

    bins = req_json("GET", f"{base}/api/v1/warehouses/{wh_id}/bins", token=token, expected=(200,)) or []
    bin_raw = find_by_code(bins, "RAW")
    if not bin_raw:
        bin_raw = req_json(
            "POST",
            f"{base}/api/v1/warehouses/{wh_id}/bins",
            token=token,
            body={"code": "RAW", "name": "Raw Materials", "bin_type": "STORAGE"},
            expected=(201,),
        )
    bin_fg = find_by_code(bins, "FG")
    if not bin_fg:
        bin_fg = req_json(
            "POST",
            f"{base}/api/v1/warehouses/{wh_id}/bins",
            token=token,
            body={"code": "FG", "name": "Finished Goods", "bin_type": "STORAGE"},
            expected=(201,),
        )

    products = req_json("GET", f"{base}/api/v1/products", token=token, expected=(200,)) or []
    needed = [
        ("RM-STEEL", "Steel Sheet"),
        ("RM-PAINT", "Paint"),
        ("FG-PUMP", "Pump Unit"),
    ]
    prod_ids: dict[str, str] = {}
    for sku, name in needed:
        p = next((x for x in products if pick(x, "sku") == sku), None)
        if not p:
            p = req_json(
                "POST",
                f"{base}/api/v1/products",
                token=token,
                body={"sku": sku, "name": name, "unit": "pcs", "tracking_mode": "NONE"},
                expected=(201,),
            )
        prod_ids[sku] = pick(p, "id")

    # Ensure at least one price per demo product,
    # otherwise /products/:id/prices returns empty and UI looks unconfigured.
    price_seed = {
        "RM-STEEL": ("PURCHASE", "300"),
        "RM-PAINT": ("PURCHASE", "90"),
        "FG-PUMP": ("SALE", "1500"),
    }
    for sku, (price_type, price) in price_seed.items():
        pid = prod_ids[sku]
        existing = req_json(
            "GET",
            f"{base}/api/v1/products/{pid}/prices",
            token=token,
            expected=(200,),
        ) or []
        has_type = any(
            str(pick(row, "price_type")).upper() == price_type for row in existing
        )
        if not has_type:
            req_json(
                "POST",
                f"{base}/api/v1/products/{pid}/prices",
                token=token,
                body={
                    "price_type": price_type,
                    "currency": "RUB",
                    "price": price,
                    "valid_from": "2026-01-01",
                },
                expected=(201,),
            )

    req_json(
        "POST",
        f"{base}/api/v1/operations/receipt",
        token=token,
        body={
            "warehouse_id": wh_id,
            "bin_id": pick(bin_raw, "id"),
            "lines": [
                {"product_id": prod_ids["RM-STEEL"], "qty": "3000", "unit_cost": "300"},
                {"product_id": prod_ids["RM-PAINT"], "qty": "2000", "unit_cost": "90"},
            ],
        },
        expected=(201,),
    )

    return {
        "warehouse_id": wh_id,
        "bin_raw_id": pick(bin_raw, "id"),
        "bin_fg_id": pick(bin_fg, "id"),
        "product_rm_steel_id": prod_ids["RM-STEEL"],
        "product_rm_paint_id": prod_ids["RM-PAINT"],
        "product_fg_pump_id": prod_ids["FG-PUMP"],
    }


def ensure_sed_catalog(base: str, token: str) -> dict[str, str]:
    wfs = req_json("GET", f"{base}/api/v1/workflows", token=token, expected=(200,)) or []
    wf = find_by_code(wfs, "WF-DEMO-APPROVAL")
    if not wf:
        wf = req_json(
            "POST",
            f"{base}/api/v1/workflows",
            token=token,
            body={"code": "WF-DEMO-APPROVAL", "name": "Demo Approval Flow"},
            expected=(201,),
        )
    wf_id = pick(wf, "id")

    steps = req_json("GET", f"{base}/api/v1/workflows/{wf_id}/steps", token=token, expected=(200,)) or []
    has_step = any(s.get("order_no") == 1 and s.get("required_role") == "sed_approver" for s in steps)
    if not has_step:
        req_json(
            "POST",
            f"{base}/api/v1/workflows/{wf_id}/steps",
            token=token,
            body={"order_no": 1, "name": "Approve", "required_role": "sed_approver"},
            expected=(201,),
        )

    dtypes = req_json("GET", f"{base}/api/v1/document-types", token=token, expected=(200,)) or []
    needed = [
        ("BOM_APPROVAL", "BOM Approval", "NONE"),
        ("ROUTING_APPROVAL", "Routing Approval", "NONE"),
        ("PURCHASE_REQUEST_APPROVAL", "PR Approval", "NONE"),
        ("PURCHASE_ORDER_APPROVAL", "PO Approval", "NONE"),
        ("MATERIAL_REQUEST", "Material Request", "RESERVE"),
    ]
    out: dict[str, str] = {}
    for code, name, action in needed:
        d = find_by_code(dtypes, code)
        if not d:
            d = req_json(
                "POST",
                f"{base}/api/v1/document-types",
                token=token,
                body={
                    "code": code,
                    "name": name,
                    "warehouse_action": action,
                    "default_workflow_id": wf_id,
                },
                expected=(201,),
            )
        out[code] = pick(d, "id")
    return out


def seed_sed_document(base: str, token: str, ids: dict[str, str], sed_type_ids: dict[str, str]) -> None:
    doc = req_json(
        "POST",
        f"{base}/api/v1/documents",
        token=token,
        body={
            "type_id": sed_type_ids["MATERIAL_REQUEST"],
            "title": f"Demo material request {now_suffix()}",
            "payload": {
                "warehouse_id": ids["warehouse_id"],
                "default_bin_id": ids["bin_raw_id"],
                "lines": [
                    {
                        "product_id": ids["product_rm_steel_id"],
                        "qty": "5",
                        "reason": "demo_seed",
                        "doc_ref": "seed-material-request",
                    }
                ],
            },
        },
        expected=(201,),
    )
    doc_id = pick(doc, "id")
    req_json("POST", f"{base}/api/v1/documents/{doc_id}/submit", token=token, expected=(204,))
    req_json(
        "POST",
        f"{base}/api/v1/documents/{doc_id}/approve",
        token=token,
        body={"comment": "seed approve"},
        expected=(204,),
    )
    req_json("POST", f"{base}/api/v1/documents/{doc_id}/sign", token=token, expected=(204,))


def seed_production(
    gateway_base: str,
    production_direct_base: str,
    token: str,
    ids: dict[str, str],
    sed_type_ids: dict[str, str],
    callback_secret: str,
) -> None:
    workcenters = req_json("GET", f"{gateway_base}/api/v1/workcenters", token=token, expected=(200,)) or []
    wc = find_by_code(workcenters, "WC-ASM")
    if not wc:
        wc = req_json(
            "POST",
            f"{gateway_base}/api/v1/workcenters",
            token=token,
            body={"code": "WC-ASM", "name": "Assembly", "active": True, "capacity_minutes_per_shift": 480},
            expected=(201,),
        )
    wc_id = pick(wc, "id")

    reasons = req_json("GET", f"{gateway_base}/api/v1/scrap-reasons", token=token, expected=(200,)) or []
    if not find_by_code(reasons, "SCRAP-DEFECT"):
        req_json(
            "POST",
            f"{gateway_base}/api/v1/scrap-reasons",
            token=token,
            body={"code": "SCRAP-DEFECT", "name": "Defect"},
            expected=(201,),
        )

    bom = req_json(
        "POST",
        f"{gateway_base}/api/v1/boms",
        token=token,
        body={"product_id": ids["product_fg_pump_id"]},
        expected=(201,),
    )
    bom_id = pick(bom, "id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/boms/{bom_id}/lines",
        token=token,
        body={
            "line_no": 1,
            "component_product_id": ids["product_rm_steel_id"],
            "qty_per": "2",
            "scrap_pct": "1.5",
            "op_no": 10,
        },
        expected=(201,),
    )
    req_json(
        "POST",
        f"{gateway_base}/api/v1/boms/{bom_id}/lines",
        token=token,
        body={
            "line_no": 2,
            "component_product_id": ids["product_rm_paint_id"],
            "qty_per": "1",
            "scrap_pct": "0",
            "op_no": 20,
        },
        expected=(201,),
    )
    req_json(
        "POST",
        f"{gateway_base}/api/v1/boms/{bom_id}/submit",
        token=token,
        body={"sed_document_type_id": sed_type_ids["BOM_APPROVAL"], "title": f"BOM seed {now_suffix()}"},
        expected=(204,),
    )

    bom_detail = req_json("GET", f"{gateway_base}/api/v1/boms/{bom_id}", token=token, expected=(200,))
    bom_doc_id = pick(bom_detail["bom"], "sed_document_id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/documents/{bom_doc_id}/approve",
        token=token,
        body={"comment": "seed approve"},
        expected=(204,),
    )
    req_json("POST", f"{gateway_base}/api/v1/documents/{bom_doc_id}/sign", token=token, expected=(204,))
    req_json(
        "POST",
        f"{production_direct_base}/api/v1/internal/sed-events",
        headers={"X-Service-Secret": callback_secret},
        body={"event": "DOCUMENT_SIGNED", "tenant_code": "devcorp", "document_id": bom_doc_id},
        expected=(204,),
    )

    routing = req_json(
        "POST",
        f"{gateway_base}/api/v1/routings",
        token=token,
        body={"product_id": ids["product_fg_pump_id"]},
        expected=(201,),
    )
    routing_id = pick(routing, "id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/routings/{routing_id}/operations",
        token=token,
        body={"op_no": 10, "workcenter_id": wc_id, "name": "Assembly", "time_per_unit_min": "15", "qc_required": True},
        expected=(201,),
    )
    req_json(
        "POST",
        f"{gateway_base}/api/v1/routings/{routing_id}/submit",
        token=token,
        body={"sed_document_type_id": sed_type_ids["ROUTING_APPROVAL"], "title": f"Routing seed {now_suffix()}"},
        expected=(204,),
    )
    routing_detail = req_json("GET", f"{gateway_base}/api/v1/routings/{routing_id}", token=token, expected=(200,))
    routing_doc_id = pick(routing_detail["routing"], "sed_document_id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/documents/{routing_doc_id}/approve",
        token=token,
        body={"comment": "seed approve"},
        expected=(204,),
    )
    req_json("POST", f"{gateway_base}/api/v1/documents/{routing_doc_id}/sign", token=token, expected=(204,))
    req_json(
        "POST",
        f"{production_direct_base}/api/v1/internal/sed-events",
        headers={"X-Service-Secret": callback_secret},
        body={"event": "DOCUMENT_SIGNED", "tenant_code": "devcorp", "document_id": routing_doc_id},
        expected=(204,),
    )

    order = req_json(
        "POST",
        f"{gateway_base}/api/v1/orders",
        token=token,
        body={
            "code": f"DEMO-PO-{now_suffix()}",
            "product_id": ids["product_fg_pump_id"],
            "bom_id": bom_id,
            "routing_id": routing_id,
            "warehouse_id": ids["warehouse_id"],
            "default_bin_id": ids["bin_raw_id"],
            "qty_planned": "10",
        },
        expected=(201,),
    )
    order_id = pick(order, "id")
    req_json("POST", f"{gateway_base}/api/v1/orders/{order_id}/release", token=token, expected=(204,))

    detail = req_json("GET", f"{gateway_base}/api/v1/orders/{order_id}", token=token, expected=(200,))
    ops = detail.get("operations") or []
    if ops:
        op_id = pick(ops[0], "id")
        req_json(
            "POST",
            f"{gateway_base}/api/v1/shift-tasks",
            token=token,
            body={
                "order_operation_id": op_id,
                "shift_date": dt.datetime.utcnow().strftime("%Y-%m-%dT00:00:00Z"),
                "shift_no": 1,
                "qty_planned": "10",
            },
            expected=(201,),
        )


def seed_procurement(
    gateway_base: str,
    procurement_direct_base: str,
    token: str,
    ids: dict[str, str],
    sed_type_ids: dict[str, str],
    callback_secret: str,
) -> None:
    suppliers = req_json("GET", f"{gateway_base}/api/v1/suppliers", token=token, expected=(200,)) or []
    supplier = find_by_code(suppliers, "SUP-DEMO")
    if not supplier:
        supplier = req_json(
            "POST",
            f"{gateway_base}/api/v1/suppliers",
            token=token,
            body={"code": "SUP-DEMO", "name": "Demo Supplier", "active": True, "contacts": {"phone": "+7-999-000-00-00"}},
            expected=(201,),
        )
    supplier_id = pick(supplier, "id")

    pr = req_json("POST", f"{gateway_base}/api/v1/purchase-requests", token=token, body={}, expected=(201,))
    pr_id = pick(pr, "id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/purchase-requests/{pr_id}/lines",
        token=token,
        body={
            "line_no": 1,
            "product_id": ids["product_rm_paint_id"],
            "qty": "25",
            "uom": "pcs",
            "target_warehouse_id": ids["warehouse_id"],
            "target_bin_id": ids["bin_raw_id"],
        },
        expected=(201,),
    )
    req_json(
        "POST",
        f"{gateway_base}/api/v1/purchase-requests/{pr_id}/submit",
        token=token,
        body={"sed_document_type_id": sed_type_ids["PURCHASE_REQUEST_APPROVAL"], "title": f"PR seed {now_suffix()}"},
        expected=(204,),
    )
    pr_detail = req_json("GET", f"{gateway_base}/api/v1/purchase-requests/{pr_id}", token=token, expected=(200,))
    pr_doc_id = pick(pr_detail["pr"], "sed_document_id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/documents/{pr_doc_id}/approve",
        token=token,
        body={"comment": "seed approve"},
        expected=(204,),
    )
    req_json("POST", f"{gateway_base}/api/v1/documents/{pr_doc_id}/sign", token=token, expected=(204,))
    req_json(
        "POST",
        f"{procurement_direct_base}/api/v1/internal/sed-events",
        headers={"X-Service-Secret": callback_secret},
        body={"event": "DOCUMENT_SIGNED", "tenant_code": "devcorp", "document_id": pr_doc_id},
        expected=(204,),
    )

    po = req_json(
        "POST",
        f"{gateway_base}/api/v1/purchase-orders/from-pr/{pr_id}",
        token=token,
        body={"supplier_id": supplier_id},
        expected=(201,),
    )
    po_id = pick(po, "id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/purchase-orders/{po_id}/submit",
        token=token,
        body={"sed_document_type_id": sed_type_ids["PURCHASE_ORDER_APPROVAL"], "title": f"PO seed {now_suffix()}"},
        expected=(204,),
    )
    po_detail = req_json("GET", f"{gateway_base}/api/v1/purchase-orders/{po_id}", token=token, expected=(200,))
    po_doc_id = pick(po_detail["po"], "sed_document_id")
    req_json(
        "POST",
        f"{gateway_base}/api/v1/documents/{po_doc_id}/approve",
        token=token,
        body={"comment": "seed approve"},
        expected=(204,),
    )
    req_json("POST", f"{gateway_base}/api/v1/documents/{po_doc_id}/sign", token=token, expected=(204,))
    req_json(
        "POST",
        f"{procurement_direct_base}/api/v1/internal/sed-events",
        headers={"X-Service-Secret": callback_secret},
        body={"event": "DOCUMENT_SIGNED", "tenant_code": "devcorp", "document_id": po_doc_id},
        expected=(204,),
    )
    req_json("POST", f"{gateway_base}/api/v1/purchase-orders/{po_id}/release", token=token, expected=(204,))
    req_json(
        "POST",
        f"{gateway_base}/api/v1/purchase-orders/{po_id}/receive",
        token=token,
        body={"warehouse_id": ids["warehouse_id"], "bin_id": ids["bin_raw_id"]},
        expected=(200,),
    )


def parse_args() -> Cfg:
    p = argparse.ArgumentParser(description="Seed demo data into VKR services.")
    p.add_argument("--gateway-base", default="http://localhost:8080")
    p.add_argument("--production-direct-base", default="http://localhost:8092")
    p.add_argument("--procurement-direct-base", default="http://localhost:8093")
    p.add_argument("--test-secret", default="test-secret-change-me")
    p.add_argument("--callback-secret", default="dev-secret")
    p.add_argument("--superadmin-username", default="superadmin")
    p.add_argument("--superadmin-password", default="superadmin")
    a = p.parse_args()
    return Cfg(
        gateway_base=a.gateway_base.rstrip("/"),
        production_direct_base=a.production_direct_base.rstrip("/"),
        procurement_direct_base=a.procurement_direct_base.rstrip("/"),
        test_secret=a.test_secret,
        callback_secret=a.callback_secret,
        superadmin_username=a.superadmin_username,
        superadmin_password=a.superadmin_password,
    )


def main() -> int:
    cfg = parse_args()
    token = login_superadmin(cfg)
    ids = ensure_warehouse_basics(cfg.gateway_base, token)
    sed_type_ids = ensure_sed_catalog(cfg.gateway_base, token)
    seed_sed_document(cfg.gateway_base, token, ids, sed_type_ids)
    seed_production(
        cfg.gateway_base,
        cfg.production_direct_base,
        token,
        ids,
        sed_type_ids,
        cfg.callback_secret,
    )
    seed_procurement(
        cfg.gateway_base,
        cfg.procurement_direct_base,
        token,
        ids,
        sed_type_ids,
        cfg.callback_secret,
    )

    balances = req_json("GET", f"{cfg.gateway_base}/api/v1/balances", token=token, expected=(200,))
    orders = req_json("GET", f"{cfg.gateway_base}/api/v1/orders", token=token, expected=(200,))
    prs = req_json("GET", f"{cfg.gateway_base}/api/v1/purchase-requests", token=token, expected=(200,))
    docs = req_json("GET", f"{cfg.gateway_base}/api/v1/documents", token=token, expected=(200,))
    print("Seed completed.")
    print(f"balances: {len(balances or [])}")
    print(f"production orders: {len(orders or [])}")
    print(f"purchase requests: {len(prs or [])}")
    print(f"documents: {len(docs or [])}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except ApiError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        raise SystemExit(1)
