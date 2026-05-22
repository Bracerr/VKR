"""Реалистичные business-payload и rag_content для RAG fixtures."""

from __future__ import annotations

import json
from typing import Any

SUPPLIERS = [
    ("ООО «ТехСнаб»", "ИНН 7701234567", "Москва"),
    ("АО «ПромКомплект»", "ИНН 5029876543", "Санкт-Петербург"),
    ("ООО «Метиз-Центр»", "ИНН 5401122334", "Новосибирск"),
]

CUSTOMERS = [
    ("ООО «Завод Север»", "ИНН 7809988776"),
    ("ПАО «АвтоДеталь»", "ИНН 7705544332"),
    ("ООО «МашКомплект»", "ИНН 5256012345"),
]

PRODUCTS = [
    ("BRG-6205-2RS", "Подшипник 6205-2RS", "шт"),
    ("BLT-M10-40", "Болт М10×40 оц.", "шт"),
    ("GSK-45-80", "Прокладка 45×80 NBR", "шт"),
    ("OIL-HLP46-20", "Масло гидравлич. HLP 46, 20л", "кан"),
    ("MTR-3KW-1500", "Двигатель 3 кВт 1500 об/мин", "шт"),
]


def _money(v: float) -> float:
    return round(v, 2)


def business_payload(code: str, slug: str, idx: int, wh_ids: dict[str, str] | None) -> tuple[str, dict[str, Any]]:
    """Возвращает (title, payload) для пользовательского API СЭД."""
    rag_ref = f"RAG-{slug}-{idx:03d}"
    sup = SUPPLIERS[(idx - 1) % len(SUPPLIERS)]
    cust = CUSTOMERS[(idx - 1) % len(CUSTOMERS)]
    prod = PRODUCTS[(idx - 1) % len(PRODUCTS)]

    if code == "PURCHASE_REQUEST_APPROVAL":
        qty, price = 80 + idx * 10, 120.0 + idx * 15
        total = _money(qty * price)
        title = f"Заявка на закупку {rag_ref} — {prod[1]}"
        payload = {
            "document_kind": "purchase_request",
            "request_no": rag_ref,
            "department": "Отдел снабжения",
            "requested_by": "Иванов А.С.",
            "needed_by": f"2026-0{min(6 + (idx % 4), 9)}-15",
            "priority": "normal" if idx < 4 else "high",
            "justification": (
                f"Плановое пополнение запасов: {prod[1]}. "
                f"Остаток на складе ниже минимума, линия сборки №{(idx % 3) + 1}."
            ),
            "cost_center": f"CC-PROD-0{idx}",
            "currency": "RUB",
            "lines": [
                {
                    "line_no": 1,
                    "sku": prod[0],
                    "name": prod[1],
                    "qty": qty,
                    "unit": prod[2],
                    "unit_price": price,
                    "amount": total,
                }
            ],
            "total_amount": total,
            "preferred_supplier": sup[0],
        }
        return title, payload

    if code == "PURCHASE_ORDER_APPROVAL":
        total = _money(45000 + idx * 8200)
        title = f"Заказ поставщику {rag_ref} — {sup[0]}"
        payload = {
            "document_kind": "purchase_order",
            "order_no": rag_ref,
            "supplier": {"name": sup[0], "inn": sup[1], "city": sup[2]},
            "payment_terms": "30 дней после поставки",
            "delivery_date": f"2026-07-{10 + idx:02d}",
            "currency": "RUB",
            "lines": [
                {
                    "line_no": 1,
                    "sku": prod[0],
                    "name": prod[1],
                    "qty": 100 + idx * 5,
                    "unit": prod[2],
                    "unit_price": 95.0 + idx,
                    "amount": total,
                }
            ],
            "total_amount": total,
            "incoterms": "DAP",
            "warehouse_destination": "Центральный склад",
        }
        return title, payload

    if code == "SUPPLIER_CONTRACT_APPROVAL":
        title = f"Договор поставки {rag_ref} — {sup[0]}"
        payload = {
            "document_kind": "supplier_contract",
            "contract_no": rag_ref,
            "supplier": {"name": sup[0], "inn": sup[1]},
            "subject": f"Поставка {prod[1]} и сопутствующих ТМЦ",
            "valid_from": "2026-01-01",
            "valid_to": "2026-12-31",
            "max_amount": _money(2_500_000 + idx * 100_000),
            "currency": "RUB",
            "penalty_percent": 0.1,
            "payment_terms": "предоплата 30%, остаток по факту поставки",
            "contact_person": "Петрова Е.В.",
        }
        return title, payload

    if code == "SALES_ORDER_APPROVAL":
        total = _money(180000 + idx * 12000)
        title = f"Заказ клиента {rag_ref} — {cust[0]}"
        payload = {
            "document_kind": "sales_order",
            "order_no": rag_ref,
            "customer": {"name": cust[0], "inn": cust[1]},
            "ship_to": "г. Москва, ул. Промышленная, 12",
            "requested_ship_date": f"2026-08-{5 + idx:02d}",
            "currency": "RUB",
            "lines": [
                {
                    "line_no": 1,
                    "sku": prod[0],
                    "name": prod[1],
                    "qty": 20 + idx * 2,
                    "unit": prod[2],
                    "unit_price": 2100.0 + idx * 50,
                    "amount": total,
                }
            ],
            "total_amount": total,
            "manager": "Сидорова М.И.",
        }
        return title, payload

    if code == "SHIPMENT_APPROVAL":
        title = f"Отгрузка {rag_ref} — {cust[0]}"
        payload = {
            "document_kind": "shipment",
            "shipment_no": rag_ref,
            "customer": cust[0],
            "sales_order_ref": f"RAG-SO-{idx:03d}",
            "carrier": "Деловые Линии",
            "planned_departure": f"2026-08-{12 + idx:02d}T09:00:00+03:00",
            "packages": 3 + idx,
            "weight_kg": _money(420.5 + idx * 18.2),
            "lines": [{"sku": prod[0], "name": prod[1], "qty": 10 + idx}],
        }
        return title, payload

    if code == "BOM_APPROVAL":
        title = f"Спецификация BOM {rag_ref} — изделие Z-{100 + idx}"
        payload = {
            "document_kind": "bom",
            "bom_no": rag_ref,
            "product_code": f"Z-{100 + idx}",
            "product_name": f"Узел приводной Z-{100 + idx}",
            "revision": f"R{idx}",
            "effective_from": "2026-06-01",
            "components": [
                {"sku": PRODUCTS[0][0], "name": PRODUCTS[0][1], "qty_per_unit": 2},
                {"sku": PRODUCTS[1][0], "name": PRODUCTS[1][1], "qty_per_unit": 8},
                {"sku": PRODUCTS[2][0], "name": PRODUCTS[2][1], "qty_per_unit": 1},
            ],
        }
        return title, payload

    if code == "ROUTING_APPROVAL":
        title = f"Маршрут производства {rag_ref} — Z-{100 + idx}"
        payload = {
            "document_kind": "routing",
            "routing_no": rag_ref,
            "product_code": f"Z-{100 + idx}",
            "operations": [
                {"op_no": 10, "workcenter": "WC-Lathe-01", "name": "Токарная обработка", "std_minutes": 45},
                {"op_no": 20, "workcenter": "WC-Mill-02", "name": "Фрезерование", "std_minutes": 30},
                {"op_no": 30, "workcenter": "WC-Assembly-01", "name": "Сборка", "std_minutes": 25},
            ],
        }
        return title, payload

    # warehouse types
    base_wh: dict[str, Any] = {}
    if wh_ids:
        base_wh = {
            "warehouse_id": wh_ids.get("warehouse_id"),
            "default_bin_id": wh_ids.get("bin_id"),
            "lines": [
                {
                    "product_id": wh_ids.get("product_id"),
                    "sku": prod[0],
                    "name": prod[1],
                    "qty": str(2 + idx),
                    "reason": rag_ref,
                    "doc_ref": rag_ref,
                }
            ],
        }
    if code == "RAG_WH_RESERVE":
        title = f"Резервирование ТМЦ {rag_ref}"
        payload = {"document_kind": "warehouse_reserve", "reserve_no": rag_ref, **base_wh}
        return title, payload
    if code == "RAG_WH_CONSUME":
        title = f"Списание в производство {rag_ref}"
        payload = {"document_kind": "warehouse_consume", "consume_no": rag_ref, **base_wh}
        return title, payload
    title = f"Оприходование на склад {rag_ref}"
    payload = {"document_kind": "warehouse_receipt", "receipt_no": rag_ref, **base_wh}
    return title, payload


def rag_content(code: str, slug: str, idx: int, title: str, payload: dict[str, Any]) -> dict[str, Any]:
    """Контент только для internal RAG API (не отдаётся в GET /documents)."""
    rag_id = f"RAG-{slug}-{idx:03d}"
    parts = [title, payload.get("justification") or payload.get("subject") or payload.get("document_kind", "")]
    if lines := payload.get("lines"):
        if isinstance(lines, list) and lines:
            parts.append(json.dumps(lines[0], ensure_ascii=False))
    if comps := payload.get("components"):
        parts.append(" ".join(c.get("name", "") for c in comps if isinstance(c, dict)))
    search_text = "\n".join(p for p in parts if p)
    return {
        "rag_id": rag_id,
        "document_kind": payload.get("document_kind", code.lower()),
        "search_text": search_text,
        "summary": search_text[:500],
        "keywords": [slug.lower(), payload.get("document_kind", code), f"fixture-{idx}"],
    }


def rag_headers(cfg_tenant: str, secret: str) -> dict[str, str]:
    return {
        "X-Service-Secret": secret,
        "X-Tenant-Id": cfg_tenant,
        "Content-Type": "application/json",
    }
