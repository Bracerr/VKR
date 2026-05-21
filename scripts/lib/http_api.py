"""Shared HTTP helpers for VKR seed scripts."""

from __future__ import annotations

import json
import urllib.error
import urllib.request
from typing import Any


class ApiError(Exception):
    pass


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
        with urllib.request.urlopen(req, timeout=120) as resp:
            status = resp.getcode()
            payload = resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        text = e.read().decode("utf-8", errors="replace")
        if e.code in expected:
            if not text:
                return None
            try:
                return json.loads(text)
            except json.JSONDecodeError:
                return text
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


def pick(item: dict[str, Any], *keys: str) -> Any:
    def norm(s: str) -> str:
        return s.lower().replace("_", "").replace("-", "")

    lowered = {norm(k): v for k, v in item.items()}
    for k in keys:
        nk = norm(k)
        if nk in lowered:
            return lowered[nk]
    raise KeyError(keys[0] if keys else "key")


def find_by_code(items: list[dict[str, Any]], code: str) -> dict[str, Any] | None:
    for item in items:
        if pick(item, "code") == code:
            return item
    return None


def login_test(auth_base: str, test_secret: str, username: str, password: str) -> str:
    data = req_json(
        "POST",
        f"{auth_base}/api/v1/internal/test/login",
        body={"username": username, "password": password},
        headers={"X-Test-Secret": test_secret},
        expected=(200,),
    )
    return data["access_token"]
