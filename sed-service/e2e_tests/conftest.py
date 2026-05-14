"""Ожидание готовности стенда (docker-compose.test.yml)."""
import os
import time

import pytest
import requests

AUTH_URL = os.environ.get("AUTH_BASE_URL", "http://localhost:28080")
WH_URL = os.environ.get("WAREHOUSE_BASE_URL", "http://localhost:28090")
SED_URL = os.environ.get("SED_BASE_URL", "http://localhost:28091")
PROD_URL = os.environ.get("PRODUCTION_BASE_URL", "http://localhost:28092")
PROC_URL = os.environ.get("PROCUREMENT_BASE_URL", "http://localhost:28093")
SALES_URL = os.environ.get("SALES_BASE_URL", "http://localhost:28094")
TRACE_URL = os.environ.get("TRACE_BASE_URL", "http://localhost:28095")

TEST_SECRET = os.environ.get("TEST_SECRET", "e2e-test-secret")
WH_SERVICE_SECRET = os.environ.get("WAREHOUSE_SERVICE_SECRET", "sed-e2e-wh-secret")


def wait_url(name: str, url: str, timeout: float = 240) -> None:
    deadline = time.time() + timeout
    last = None
    while time.time() < deadline:
        try:
            r = requests.get(url, timeout=5)
            if r.status_code == 200:
                return
            last = r.text
        except Exception as e:
            last = str(e)
        time.sleep(2)
    pytest.fail(f"{name} not ready: {last}")


@pytest.fixture(scope="session", autouse=True)
def wait_stack():
    wait_url("auth", f"{AUTH_URL}/ready")
    wait_url("warehouse", f"{WH_URL}/health")
    wait_url("sed", f"{SED_URL}/health")
    wait_url("production", f"{PROD_URL}/health")
    wait_url("procurement", f"{PROC_URL}/health")
    wait_url("sales", f"{SALES_URL}/health")
    wait_url("traceability", f"{TRACE_URL}/health")
    try:
        requests.delete(
            f"{AUTH_URL}/api/v1/internal/test/cleanup",
            headers={"X-Test-Secret": TEST_SECRET},
            timeout=60,
        )
    except Exception:
        pass


@pytest.fixture
def auth_api():
    return AUTH_URL


@pytest.fixture
def wh_api():
    return WH_URL


@pytest.fixture
def sed_api():
    return SED_URL


@pytest.fixture
def prod_api():
    return PROD_URL


@pytest.fixture
def proc_api():
    return PROC_URL


@pytest.fixture
def sales_api():
    return SALES_URL


@pytest.fixture
def trace_api():
    return TRACE_URL


@pytest.fixture
def headers_test():
    return {"X-Test-Secret": TEST_SECRET, "Content-Type": "application/json"}


@pytest.fixture
def wh_svc_headers():
    def _h(tenant: str):
        return {
            "X-Service-Secret": WH_SERVICE_SECRET,
            "X-Tenant-Id": tenant,
            "Content-Type": "application/json",
        }

    return _h
