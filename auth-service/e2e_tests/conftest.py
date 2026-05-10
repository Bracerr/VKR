"""Фикстуры pytest: ожидание готовности auth-service."""
import os
import time

import pytest
import requests

BASE_URL = os.environ.get("AUTH_BASE_URL", "http://localhost:18080")
TEST_SECRET = os.environ.get("TEST_SECRET", "e2e-test-secret")
SERVICE_SECRET = os.environ.get("SERVICE_SECRET", "e2e-service-secret")


@pytest.fixture(scope="session", autouse=True)
def wait_ready():
    """Ждём /ready (PostgreSQL + Keycloak realm)."""
    deadline = time.time() + 180
    last_err = None
    while time.time() < deadline:
        try:
            r = requests.get(f"{BASE_URL}/ready", timeout=5)
            if r.status_code == 200:
                try:
                    requests.delete(
                        f"{BASE_URL}/api/v1/internal/test/cleanup",
                        headers={"X-Test-Secret": TEST_SECRET},
                        timeout=30,
                    )
                except Exception:
                    pass
                return
            last_err = r.text
        except Exception as e:
            last_err = str(e)
        time.sleep(2)
    pytest.fail(f"auth-service not ready: {last_err}")


@pytest.fixture
def api():
    return BASE_URL


@pytest.fixture
def headers_test():
    return {"X-Test-Secret": TEST_SECRET, "Content-Type": "application/json"}


@pytest.fixture
def headers_service():
    return {"X-Service-Secret": SERVICE_SECRET}
